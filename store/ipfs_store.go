package store

import (
	"context"
	"io"
	"net/http"

	api "github.com/ipfs/go-ipfs-api"
)

type IpfsStore struct {
	shell *api.Shell
}

func NewIpfsStore(url string) IpfsStore {
	shell := api.NewShell(url)
	return IpfsStore{
		shell: shell,
	}
}

func NewIpfsStoreWithBasicAuth(url string, username string, password string) IpfsStore {
	shell := api.NewShellWithClient(url, NewClient(username, password))
	return IpfsStore{
		shell: shell,
	}
}

// NewClient creates an http.Client that automatically perform basic auth on each request.
func NewClient(projectId, projectSecret string) *http.Client {
	return &http.Client{
		Transport: authTransport{
			RoundTripper:  http.DefaultTransport,
			ProjectId:     projectId,
			ProjectSecret: projectSecret,
		},
	}
}

// authTransport decorates each request with a basic auth header.
type authTransport struct {
	http.RoundTripper
	ProjectId     string
	ProjectSecret string
}

func (t authTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.SetBasicAuth(t.ProjectId, t.ProjectSecret)
	return t.RoundTripper.RoundTrip(r)
}

func (s IpfsStore) StoreFile(ctx context.Context, reader io.Reader, info map[string]string) (StoreRet, error) {
	hash, err := s.shell.Add(reader)
	if err != nil {
		return StoreRet{}, err
	}

	return StoreRet{
		IpfsHash: hash,
	}, nil
}

func (s IpfsStore) DeleteFile(ctx context.Context, info map[string]string) error {
	return s.shell.Unpin(info["hash"])
}

func (s IpfsStore) GetFile(ctx context.Context, info map[string]string) (io.ReadCloser, error) {
	return s.shell.Cat(info["hash"])
}
