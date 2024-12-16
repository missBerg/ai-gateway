package translator

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	extprocv3http "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_proc/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/ai-gateway/internal/apischema/openai"
)

func TestNewOpenAIToOpenAITranslator(t *testing.T) {
	t.Run("unsupported path", func(t *testing.T) {
		_, err := newOpenAIToOpenAITranslator("/v1/foo/bar", slog.Default())
		require.Error(t, err)
	})
	t.Run("/v1/chat/completions", func(t *testing.T) {
		translator, err := newOpenAIToOpenAITranslator("/v1/chat/completions", slog.Default())
		require.NoError(t, err)
		require.NotNil(t, translator)
	})
}

func TestOpenAIToOpenAITranslatorV1ChatCompletionRequestBody(t *testing.T) {
	t.Run("invalid body", func(t *testing.T) {
		o := &openAIToOpenAITranslatorV1ChatCompletion{}
		_, _, _, _, err := o.RequestBody(&extprocv3.HttpBody{Body: []byte("invalid")})
		require.Error(t, err)
	})
	t.Run("valid body", func(t *testing.T) {
		for _, stream := range []bool{true, false} {
			t.Run(fmt.Sprintf("stream=%t", stream), func(t *testing.T) {
				originalReq := openai.ChatCompletionRequest{Model: "foo-bar-ai", Stream: stream}
				body, err := json.Marshal(originalReq)
				require.NoError(t, err)

				o := &openAIToOpenAITranslatorV1ChatCompletion{}
				hm, bm, mode, modelName, err := o.RequestBody(&extprocv3.HttpBody{Body: body})
				require.Nil(t, bm)
				require.NoError(t, err)
				require.Equal(t, "foo-bar-ai", modelName)
				require.Equal(t, stream, o.stream)
				if stream {
					require.NotNil(t, mode)
					require.Equal(t, extprocv3http.ProcessingMode_SEND, mode.ResponseHeaderMode)
					require.Equal(t, extprocv3http.ProcessingMode_STREAMED, mode.ResponseBodyMode)
				} else {
					require.Nil(t, mode)
				}

				require.Nil(t, hm)
			})
		}
	})
}

func TestOpenAIToOpenAITranslatorV1ChatCompletionResponseBody(t *testing.T) {
	t.Run("streaming", func(t *testing.T) {
		// This is the real event stream from OpenAI.
		wholeBody := []byte(`
data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"role":"assistant","content":"","refusal":null},"logprobs":null,"finish_reason":null}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"content":"This"},"logprobs":null,"finish_reason":null}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"content":" is"},"logprobs":null,"finish_reason":null}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"content":" a"},"logprobs":null,"finish_reason":null}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"content":" test"},"logprobs":null,"finish_reason":null}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"content":"!"},"logprobs":null,"finish_reason":null}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"content":" How"},"logprobs":null,"finish_reason":null}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"content":" can"},"logprobs":null,"finish_reason":null}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"content":" I"},"logprobs":null,"finish_reason":null}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"content":" assist"},"logprobs":null,"finish_reason":null}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"content":" you"},"logprobs":null,"finish_reason":null}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"content":" today"},"logprobs":null,"finish_reason":null}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{"content":"?"},"logprobs":null,"finish_reason":null}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[{"index":0,"delta":{},"logprobs":null,"finish_reason":"stop"}],"usage":null}

data: {"id":"chatcmpl-foo","object":"chat.completion.chunk","created":1731618222,"model":"gpt-4o-mini-2024-07-18","system_fingerprint":"fp_0ba0d124f1","choices":[],"usage":{"prompt_tokens":13,"completion_tokens":12,"total_tokens":25,"prompt_tokens_details":{"cached_tokens":0,"audio_tokens":0},"completion_tokens_details":{"reasoning_tokens":0,"audio_tokens":0,"accepted_prediction_tokens":0,"rejected_prediction_tokens":0}}}

data: [DONE]

`)

		o := &openAIToOpenAITranslatorV1ChatCompletion{stream: true, l: slog.Default()}
		var usedToken uint32
		for i := 0; i < len(wholeBody); i++ {
			hm, bm, _usedToken, err := o.ResponseBody(&extprocv3.HttpBody{Body: wholeBody[i : i+1]})
			require.NoError(t, err)
			require.Nil(t, hm)
			require.Nil(t, bm)
			if _usedToken > 0 {
				usedToken = _usedToken
			}
			if usedToken > 0 {
				require.True(t, o.bufferingDone)
			} else {
				require.False(t, o.bufferingDone)
			}
		}
		require.Equal(t, uint32(25), usedToken)
	})
	t.Run("non-streaming", func(t *testing.T) {
		t.Run("invalid body", func(t *testing.T) {
			o := &openAIToOpenAITranslatorV1ChatCompletion{}
			_, _, _, err := o.ResponseBody(&extprocv3.HttpBody{Body: []byte("invalid")})
			require.Error(t, err)
		})
		t.Run("valid body", func(t *testing.T) {
			var resp openai.ChatCompletionResponse
			resp.Usage.TotalTokens = 42
			body, err := json.Marshal(resp)
			require.NoError(t, err)
			o := &openAIToOpenAITranslatorV1ChatCompletion{}
			_, _, usedToken, err := o.ResponseBody(&extprocv3.HttpBody{Body: body})
			require.NoError(t, err)
			require.Equal(t, uint32(42), usedToken)
		})
	})
}

func TestExtractUsageFromBufferEvent(t *testing.T) {
	logger := slog.Default()

	t.Run("valid usage data", func(t *testing.T) {
		o := &openAIToOpenAITranslatorV1ChatCompletion{l: logger}
		o.buffered = []byte("data: {\"usage\": {\"total_tokens\": 42}}\n")
		usedToken := o.extractUsageFromBufferEvent()
		require.Equal(t, uint32(42), usedToken)
		require.True(t, o.bufferingDone)
		require.Nil(t, o.buffered)
	})

	t.Run("valid usage data after invalid", func(t *testing.T) {
		o := &openAIToOpenAITranslatorV1ChatCompletion{l: logger}
		o.buffered = []byte("data: invalid\ndata: {\"usage\": {\"total_tokens\": 42}}\n")
		usedToken := o.extractUsageFromBufferEvent()
		require.Equal(t, uint32(42), usedToken)
		require.True(t, o.bufferingDone)
		require.Nil(t, o.buffered)
	})

	t.Run("no usage data and then become valid", func(t *testing.T) {
		o := &openAIToOpenAITranslatorV1ChatCompletion{l: logger}
		o.buffered = []byte("data: {}\n\ndata: ")
		usedToken := o.extractUsageFromBufferEvent()
		require.Equal(t, uint32(0), usedToken)
		require.False(t, o.bufferingDone)
		require.NotNil(t, o.buffered)

		o.buffered = append(o.buffered, []byte("{\"usage\": {\"total_tokens\": 42}}\n")...)
		usedToken = o.extractUsageFromBufferEvent()
		require.Equal(t, uint32(42), usedToken)
		require.True(t, o.bufferingDone)
		require.Nil(t, o.buffered)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		o := &openAIToOpenAITranslatorV1ChatCompletion{l: logger}
		o.buffered = []byte("data: invalid\n")
		usedToken := o.extractUsageFromBufferEvent()
		require.Equal(t, uint32(0), usedToken)
		require.False(t, o.bufferingDone)
		require.NotNil(t, o.buffered)
	})
}