package remnawave

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type Host struct {
	UUID    string `json:"uuid"`
	Remark  string `json:"remark"`
	Address string `json:"address"`
	Port    int    `json:"port"`

	// Поддерживаем оба варианта API.
	Tag  *string         `json:"tag"`
	Tags json.RawMessage `json:"tags"`
}

type InternalSquad struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

type User struct {
	UUID                 string          `json:"uuid"`
	ShortUUID            string          `json:"shortUuid"`
	Username             string          `json:"username"`
	ActiveInternalSquads []InternalSquad `json:"activeInternalSquads"`
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,

		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetHosts(ctx context.Context) ([]Host, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"/api/hosts",
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Authorization",
		"Bearer "+c.token,
	)

	req.Header.Set(
		"Accept",
		"application/json",
	)

	log.Printf(
		"[remnawave] GET %s/api/hosts",
		c.baseURL,
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf(
		"[remnawave] response status=%s",
		resp.Status,
	)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(
			io.LimitReader(resp.Body, 4096),
		)

		return nil, fmt.Errorf(
			"Remnawave API returned %s: %s",
			resp.Status,
			string(body),
		)
	}

	var envelope struct {
		Response json.RawMessage `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}

	var hosts []Host

	if err := json.Unmarshal(envelope.Response, &hosts); err == nil {
		return hosts, nil
	}

	var wrapped struct {
		Hosts []Host `json:"hosts"`
	}

	if err := json.Unmarshal(
		envelope.Response,
		&wrapped,
	); err != nil {
		return nil, fmt.Errorf(
			"unexpected Remnawave hosts response: %w",
			err,
		)
	}

	return wrapped.Hosts, nil
}

func (c *Client) GetUserByShortUUID(
	ctx context.Context,
	shortUUID string,
) (*User, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"/api/users/by-short-uuid/"+url.PathEscape(shortUUID),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Authorization",
		"Bearer "+c.token,
	)

	req.Header.Set(
		"Accept",
		"application/json",
	)

	log.Printf(
		"[remnawave] GET user by shortUuid=%s",
		shortUUID,
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf(
		"[remnawave] user response status=%s",
		resp.Status,
	)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(
			io.LimitReader(resp.Body, 4096),
		)

		return nil, fmt.Errorf(
			"Remnawave API returned %s: %s",
			resp.Status,
			string(body),
		)
	}

	var envelope struct {
		Response json.RawMessage `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}

	var user User

	if err := json.Unmarshal(envelope.Response, &user); err == nil && user.UUID != "" {
		return &user, nil
	}

	// На случай если API оборачивает пользователя:
	// {
	//   "response": {
	//     "user": {...}
	//   }
	// }
	var wrapped struct {
		User User `json:"user"`
	}

	if err := json.Unmarshal(
		envelope.Response,
		&wrapped,
	); err != nil {
		return nil, fmt.Errorf(
			"unexpected Remnawave user response: %w",
			err,
		)
	}

	if wrapped.User.UUID == "" {
		return nil, fmt.Errorf(
			"user with shortUuid %q not found in response",
			shortUUID,
		)
	}

	return &wrapped.User, nil
}

func FindHostByRemark(
	hosts []Host,
	remark string,
) (*Host, error) {

	var found *Host

	for i := range hosts {
		if hosts[i].Remark != remark {
			continue
		}

		if found != nil {
			return nil, fmt.Errorf(
				"multiple hosts found with remark %q",
				remark,
			)
		}

		found = &hosts[i]
	}

	if found == nil {
		return nil, fmt.Errorf(
			"host with remark %q not found",
			remark,
		)
	}

	return found, nil
}

func (h Host) TagNames() []string {
	result := make([]string, 0)

	seen := make(map[string]struct{})

	add := func(tag string) {
		tag = strings.TrimSpace(tag)

		if tag == "" {
			return
		}

		if _, exists := seen[tag]; exists {
			return
		}

		seen[tag] = struct{}{}
		result = append(result, tag)
	}

	// Старый/одиночный tag.
	if h.Tag != nil {
		add(*h.Tag)
	}

	if len(h.Tags) == 0 {
		return result
	}

	// tags: ["FOREIGN", "PUBLIC"]
	var stringTags []string

	if err := json.Unmarshal(h.Tags, &stringTags); err == nil {
		for _, tag := range stringTags {
			add(tag)
		}

		return result
	}

	// На случай если API отдаёт объекты тегов.
	var objectTags []struct {
		Name  string `json:"name"`
		Tag   string `json:"tag"`
		Value string `json:"value"`
	}

	if err := json.Unmarshal(h.Tags, &objectTags); err == nil {
		for _, tag := range objectTags {
			switch {
			case tag.Name != "":
				add(tag.Name)

			case tag.Tag != "":
				add(tag.Tag)

			case tag.Value != "":
				add(tag.Value)
			}
		}
	}

	return result
}

func (u User) HasInternalSquad(name string) bool {
	for _, squad := range u.ActiveInternalSquads {
		if strings.EqualFold(
			strings.TrimSpace(squad.Name),
			strings.TrimSpace(name),
		) {
			return true
		}
	}

	return false
}

func (u User) InternalSquadNames() []string {
	result := make([]string, 0, len(u.ActiveInternalSquads))

	for _, squad := range u.ActiveInternalSquads {
		if squad.Name == "" {
			continue
		}

		result = append(
			result,
			squad.Name,
		)
	}

	return result
}