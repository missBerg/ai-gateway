// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package backendauth

import (
	"context"
	"fmt"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"

	"github.com/envoyproxy/ai-gateway/internal/filterapi"
)

// apiKeyHandler implements [Handler] for api key authz.
type apiKeyHandler struct {
	apiKey string
}

func newAPIKeyHandler(auth *filterapi.APIKeyAuth) (Handler, error) {
	return &apiKeyHandler{apiKey: strings.TrimSpace(auth.Key)}, nil
}

// Do implements [Handler.Do].
//
// Extracts the api key from the local file and set it as an authorization header.
func (a *apiKeyHandler) Do(_ context.Context, requestHeaders map[string]string, headerMut *extprocv3.HeaderMutation, _ *extprocv3.BodyMutation) error {
	requestHeaders["Authorization"] = fmt.Sprintf("Bearer %s", a.apiKey)
	headerMut.SetHeaders = append(headerMut.SetHeaders, &corev3.HeaderValueOption{
		Header: &corev3.HeaderValue{Key: "Authorization", RawValue: []byte(requestHeaders["Authorization"])},
	})
	return nil
}
