// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package extproc

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
)

// contentDecodingResult contains the result of content decoding operation.
type contentDecodingResult struct {
	reader    io.Reader
	isEncoded bool
}

// decodeContentIfNeeded decompresses the response body based on the content-encoding header.
// Currently supports gzip encoding, but can be extended to support other encodings in the future.
// Returns a reader for the (potentially decompressed) body and metadata about the encoding.
func decodeContentIfNeeded(body []byte, contentEncoding string) (contentDecodingResult, error) {
	switch contentEncoding {
	case "gzip":
		reader, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return contentDecodingResult{}, fmt.Errorf("failed to decode gzip: %w", err)
		}
		return contentDecodingResult{
			reader:    reader,
			isEncoded: true,
		}, nil
	default:
		return contentDecodingResult{
			reader:    bytes.NewReader(body),
			isEncoded: false,
		}, nil
	}
}

// removeContentEncodingIfNeeded removes the content-encoding header if the body was modified and was encoded.
// This is needed when the transformation modifies the body content but the response was originally compressed.
func removeContentEncodingIfNeeded(headerMutation *extprocv3.HeaderMutation, bodyMutation *extprocv3.BodyMutation, isEncoded bool) *extprocv3.HeaderMutation {
	if bodyMutation != nil && isEncoded {
		if headerMutation == nil {
			headerMutation = &extprocv3.HeaderMutation{}
		}
		// TODO: this is a hotfix, we should update this to recompress since its in the header
		// If the upstream response was compressed and we decompressed it,
		// ensure we remove the content-encoding header.
		//
		// This is only needed when the transformation is actually modifying the body.
		headerMutation.RemoveHeaders = append(headerMutation.RemoveHeaders, "content-encoding")
	}
	return headerMutation
}

// isGoodStatusCode checks if the HTTP status code of the upstream response is successful.
// The 2xx - Successful: The request is received by upstream and processed successfully.
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Status#successful_responses
func isGoodStatusCode(code int) bool {
	return code >= 200 && code < 300
}
