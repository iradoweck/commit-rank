package github

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"most-active-github-users-counter/net"
)

const root string = "https://api.github.com/"

var sharedHTTPClient = &http.Client{
	Timeout: 60 * time.Second,
}

type HTTPGithubClient struct {
	wrappers []net.Wrapper
}

func (client HTTPGithubClient) Request(url string, body string) ([]byte, error) {
	var req *http.Request
	var err error
	if body != "" {
		req, err = http.NewRequest("POST", url, strings.NewReader(body))
	} else {
		req, err = http.NewRequest("GET", url, nil)
	}

	if err != nil {
		return []byte{}, err
	}

	return net.Compose(client.wrappers...)(net.MakeRequester(sharedHTTPClient))(req)
}

func (client HTTPGithubClient) CurrentUser() (User, error) {
	body, err := client.Request(fmt.Sprintf("%suser", root), "")
	if err != nil {
		return User{}, err
	}

	user := User{}
	if err := json.Unmarshal(body, &user); err != nil {
		return User{}, err
	}
	return user, nil
}

func (client HTTPGithubClient) User(login string) (User, error) {
	body, err := client.Request(fmt.Sprintf("%susers/%s", root, login), "")
	if err != nil {
		return User{}, err
	}

	user := User{}
	if err := json.Unmarshal(body, &user); err != nil {
		return User{}, err
	}
	return user, nil
}

var newlineRegex = regexp.MustCompile(`\r?\n`)

func (client HTTPGithubClient) SearchUsers(query UserSearchQuery) (GithubSearchResults, error) {
	users := []User{}
	userLogins := map[string]bool{}

	minFollowerCount := -1
	maxPerQuery := 1000
	// Paginação dinâmica balanceada: 10 nós por padrão (100 páginas para 1000 devs)
	// Em queries hiperdensas (worldwide ou US) ou quando especificado em PerPage, ajusta de forma adaptativa.
	perPage := 10
	if query.PerPage > 0 {
		perPage = query.PerPage
	} else if strings.Contains(query.Q, "type:user") && (len(query.Q) < 15 || strings.Contains(query.Q, "California") || strings.Contains(query.Q, "New+York")) {
		// Presets globais ou estados densos dos EUA usam lote menor (5) para evitar timeouts da API do GitHub
		perPage = 5
	}
	totalUsersCount := 0
	totalUsersCountSet := false

	maxPageRetries := 6
	toxicSkips := 0
	maxToxicSkips := 5

Pages:
	for len(users) < query.MaxUsers {
		previousCursor := ""
		followerCountQueryStr := ""
		if minFollowerCount >= 0 {
			followerCountQueryStr = fmt.Sprintf(" followers:<%d", minFollowerCount)
		}
		for currentPage := 1; currentPage <= (maxPerQuery / perPage); currentPage++ {
			if len(users) >= query.MaxUsers {
				break Pages
			}

			cursorQueryStr := ""
			if previousCursor != "" {
				cursorQueryStr = fmt.Sprintf(", after: \\\"%s\\\"", previousCursor)
			}

			var searchNode map[string]interface{}
			var lastErr error

			for attempt := 1; attempt <= maxPageRetries; attempt++ {
				// Na primeira tentativa busca em lote de 5; em retries reduz imediatamente para 1 nó para aliviar o backend GraphQL
				currentPerPage := perPage
				if attempt >= 2 {
					currentPerPage = 1
				}

				graphQlString := fmt.Sprintf(`{ "query": "query {
        search(type: USER, query:\"%s%s sort:%s-%s\", first: %d%s) {
          userCount
          edges {
            node {
              __typename
              ... on User {
                login,
                avatarUrl,
                name,
                company,
                organizations(first: 5) {
                  nodes {
                    login
                  }
                }
                followers {
                  totalCount
                }
                contributionsCollection {
                  contributionCalendar {
                    totalContributions
                  },
                  totalCommitContributions,
                  totalPullRequestContributions,
                  restrictedContributionsCount
                }
              }
            },
            cursor
          }
        }
      }" }`, query.Q, followerCountQueryStr, query.Sort, query.Order, currentPerPage, cursorQueryStr)

				graphQlString = newlineRegex.ReplaceAllString(graphQlString, " ")

				body, err := client.Request("https://api.github.com/graphql", graphQlString)
				if err != nil {
					lastErr = fmt.Errorf("error making graphql request: %w", err)
					log.Printf("page %d (retry %d/%d): %v", currentPage, attempt, maxPageRetries, err)
					if attempt < maxPageRetries {
						retryDelay := time.Duration(attempt)*3*time.Second + time.Duration(time.Now().UnixNano()%2000)*time.Millisecond
						time.Sleep(retryDelay)
					}
					continue
				}

				var response interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					lastErr = fmt.Errorf("error unmarshalling JSON response: %w", err)
					log.Printf("page %d (retry %d/%d): %v", currentPage, attempt, maxPageRetries, lastErr)
					if attempt < maxPageRetries {
						retryDelay := time.Duration(attempt)*3*time.Second + time.Duration(time.Now().UnixNano()%2000)*time.Millisecond
						time.Sleep(retryDelay)
					}
					continue
				}

				rootNode, ok := response.(map[string]interface{})
				if !ok {
					lastErr = fmt.Errorf("unexpected GraphQL response structure")
					log.Printf("page %d (retry %d/%d): %v", currentPage, attempt, maxPageRetries, lastErr)
					if attempt < maxPageRetries {
						retryDelay := time.Duration(attempt)*3*time.Second + time.Duration(time.Now().UnixNano()%2000)*time.Millisecond
						time.Sleep(retryDelay)
					}
					continue
				}

				if val, hasErrors := rootNode["errors"]; hasErrors {
					if dataNode, ok := rootNode["data"].(map[string]interface{}); ok && dataNode != nil {
						if sn, ok := dataNode["search"].(map[string]interface{}); ok && sn != nil {
							if edges, ok := sn["edges"].([]interface{}); ok && len(edges) > 0 {
								log.Printf("page %d returned partial data with GraphQL field errors; salvaging %d edges", currentPage, len(edges))
								searchNode = sn
								lastErr = nil
								break
							}
						}
					}
					lastErr = fmt.Errorf("GraphQL error response: %+v", val)
					log.Printf("page %d (retry %d/%d): %v", currentPage, attempt, maxPageRetries, lastErr)
					if attempt < maxPageRetries {
						retryDelay := time.Duration(attempt)*3*time.Second + time.Duration(time.Now().UnixNano()%2000)*time.Millisecond
						time.Sleep(retryDelay)
					}
					continue
				}

				dataNode, ok := rootNode["data"].(map[string]interface{})
				if !ok || dataNode == nil || dataNode["search"] == nil {
					lastErr = fmt.Errorf("GraphQL response missing data.search node")
					log.Printf("page %d (retry %d/%d): %v", currentPage, attempt, maxPageRetries, lastErr)
					if attempt < maxPageRetries {
						retryDelay := time.Duration(attempt)*3*time.Second + time.Duration(time.Now().UnixNano()%2000)*time.Millisecond
						time.Sleep(retryDelay)
					}
					continue
				}

				sn, ok := dataNode["search"].(map[string]interface{})
				if !ok || sn == nil {
					lastErr = fmt.Errorf("unexpected search node structure")
					log.Printf("page %d (retry %d/%d): %v", currentPage, attempt, maxPageRetries, lastErr)
					if attempt < maxPageRetries {
						retryDelay := time.Duration(attempt)*3*time.Second + time.Duration(time.Now().UnixNano()%2000)*time.Millisecond
						time.Sleep(retryDelay)
					}
					continue
				}

				searchNode = sn
				lastErr = nil
				break
			}

			if lastErr != nil {
				errMsg := lastErr.Error()
				isServerOrRateLimit := strings.Contains(errMsg, "502") || strings.Contains(errMsg, "503") || strings.Contains(errMsg, "504") || strings.Contains(errMsg, "rate limit")

				// Degradação graciosa: se o GitHub estiver instável (502/504/rate-limit) e já temos dados, preserva os usuários coletados
				if isServerOrRateLimit && len(users) > 0 {
					log.Printf("warning: page %d encountered persistent server/rate error (%v); gracefully preserving %d users collected so far", currentPage, lastErr, len(users))
					break Pages
				}

				if minFollowerCount > 0 && toxicSkips < maxToxicSkips && !isServerOrRateLimit {
					toxicSkips++
					minFollowerCount--
					log.Printf("skipping toxic cursor via followers:<%d (skip %d/%d)", minFollowerCount, toxicSkips, maxToxicSkips)
					continue Pages
				}
				if len(users) > 0 {
					log.Printf("warning: page %d exhausted %d retries (%v); gracefully returning %d users collected so far", currentPage, maxPageRetries, lastErr, len(users))
					break Pages
				}
				return GithubSearchResults{}, fmt.Errorf("too many errors on page %d: %w", currentPage, lastErr)
			}
			toxicSkips = 0

			if !totalUsersCountSet {
				if uc, ok := searchNode["userCount"].(float64); ok {
					totalUsersCount = int(uc)
					totalUsersCountSet = true
				}
			}

			edgeNodes, ok := searchNode["edges"].([]interface{})
			if !ok || len(edgeNodes) == 0 {
				break Pages
			}

		Edges:
			for _, edge := range edgeNodes {
				edgeNode, ok := edge.(map[string]interface{})
				if !ok {
					continue Edges
				}
				if c, ok := edgeNode["cursor"].(string); ok && c != "" {
					previousCursor = c
				}
				userNode, ok := edgeNode["node"].(map[string]interface{})
				if !ok || userNode == nil {
					continue Edges
				}
				typename, _ := userNode["__typename"].(string)
				if typename != "User" {
					continue Edges
				}
				login, _ := userNode["login"].(string)
				if login == "" {
					continue Edges
				}
				avatarURL := strPropOrEmpty(userNode, "avatarUrl")
				name := strPropOrEmpty(userNode, "name")
				company := strPropOrEmpty(userNode, "company")
				organizations := []string{}

				if orgsContainer, ok := userNode["organizations"].(map[string]interface{}); ok {
					if orgNodes, ok := orgsContainer["nodes"].([]interface{}); ok {
						for _, orgNode := range orgNodes {
							if orgMap, ok := orgNode.(map[string]interface{}); ok {
								if orgLogin, ok := orgMap["login"].(string); ok && orgLogin != "" {
									organizations = append(organizations, orgLogin)
								}
							}
						}
					}
				}

				followerCount := 0
				if followersMap, ok := userNode["followers"].(map[string]interface{}); ok {
					if fc, ok := followersMap["totalCount"].(float64); ok {
						followerCount = int(fc)
					}
				}

				contributionCount := 0
				privateContributionCount := 0
				commitsCount := 0
				pullRequestsCount := 0
				if contributionsCollection, ok := userNode["contributionsCollection"].(map[string]interface{}); ok {
					if calendar, ok := contributionsCollection["contributionCalendar"].(map[string]interface{}); ok {
						if tc, ok := calendar["totalContributions"].(float64); ok {
							contributionCount = int(tc)
						}
					}
					if rc, ok := contributionsCollection["restrictedContributionsCount"].(float64); ok {
						privateContributionCount = int(rc)
					}
					if cc, ok := contributionsCollection["totalCommitContributions"].(float64); ok {
						commitsCount = int(cc)
					}
					if prc, ok := contributionsCollection["totalPullRequestContributions"].(float64); ok {
						pullRequestsCount = int(prc)
					}
				}

				user := User{
					Login:                    login,
					AvatarURL:                avatarURL,
					Name:                     name,
					Company:                  company,
					Organizations:            organizations,
					FollowerCount:            followerCount,
					ContributionCount:        contributionCount,
					PublicContributionCount:  (contributionCount - privateContributionCount),
					PrivateContributionCount: privateContributionCount,
					CommitsCount:             commitsCount,
					PullRequestsCount:        pullRequestsCount}

				if !userLogins[login] {
					userLogins[login] = true
					users = append(users, user)
				}

				minFollowerCount = followerCount
				if len(users) >= query.MaxUsers {
					break Pages
				}
			}

			// Intervalo anti-rate limit com jitter suave entre páginas
			time.Sleep(420*time.Millisecond + time.Duration(time.Now().UnixNano()%100)*time.Millisecond)
		}
	}

	return GithubSearchResults{
		Users:                users,
		MinimumFollowerCount: minFollowerCount,
		TotalUserCount:       totalUsersCount}, nil
}

func strPropOrEmpty(obj map[string]interface{}, prop string) string {
	switch t := obj[prop].(type) {
	case string:
		return t
	default:
		return ""
	}
}

func (client HTTPGithubClient) Organizations(login string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/orgs", login)
	body, err := client.Request(url, "")
	if err != nil {
		return []string{}, fmt.Errorf("error requesting organizations for user %s: %w", login, err)
	}
	orgResp := []OrgResponse{}
	err = json.Unmarshal(body, &orgResp)
	if err != nil {
		return []string{}, fmt.Errorf("error parsing organizations JSON for user %s: %w", login, err)
	}
	orgs := []string{}

	for _, org := range orgResp {
		orgs = append(orgs, org.Organization)
	}

	return orgs, nil
}

type OrgResponse struct {
	Organization string `json:"login"`
}

func NewGithubClient(wrappers ...net.Wrapper) HTTPGithubClient {
	return HTTPGithubClient{wrappers: wrappers}
}

type User struct {
	Login                    string
	AvatarURL                string
	Name                     string
	Company                  string
	Organizations            []string
	FollowerCount            int
	ContributionCount        int
	PublicContributionCount  int
	PrivateContributionCount int
	CommitsCount             int
	PullRequestsCount        int
}

type UserSearchQuery struct {
	Q        string
	Sort     string
	Order    string
	MaxUsers int
	PerPage  int
}

type GithubSearchResults struct {
	Users                []User
	MinimumFollowerCount int
	TotalUserCount       int
}
