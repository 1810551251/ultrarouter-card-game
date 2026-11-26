package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Session-only handler coverage lives in model/gacha_game_test.go. These
// database-free cases cover transport parsing not exercised by that matrix.
func TestGachaIntentDecoder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("finish_accepts_only_intent_values", func(t *testing.T) {
		for _, tc := range []struct {
			name    string
			payload string
			abandon bool
		}{
			{"claim", `{"run_id":"run_server-issued"}`, false},
			{"explicit_claim", `{"run_id":"run_server-issued","abandon":false}`, false},
			{"abandon", `{"run_id":"run_server-issued","abandon":true}`, true},
		} {
			t.Run(tc.name, func(t *testing.T) {
				recorder := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest(http.MethodPost, "/api/user/gacha/run/finish", strings.NewReader(tc.payload))
				var intent struct {
					RunID   string `json:"run_id"`
					Abandon bool   `json:"abandon"`
				}
				require.True(t, decodeGachaIntent(c, &intent, []string{"run_id", "abandon"}))
				assert.Equal(t, "run_server-issued", intent.RunID)
				assert.Equal(t, tc.abandon, intent.Abandon)
				assert.Empty(t, recorder.Body.String(), "valid decoding does not write an API response or touch the database")
			})
		}
	})

	t.Run("rejects_ambiguous_or_unbounded_transport", func(t *testing.T) {
		for _, tc := range []struct {
			name    string
			payload string
		}{
			{"unknown_params", `{"run_id":"r","params":{"stars":3}}`},
			{"unknown_capitalized_key", `{"run_id":"r","Abandon":true}`},
			{"object_run_id", `{"run_id":{"id":"r"}}`},
			{"numeric_run_id", `{"run_id":1}`},
			{"string_boolean", `{"run_id":"r","abandon":"false"}`},
			{"numeric_boolean", `{"run_id":"r","abandon":0}`},
			{"array_instead_of_object", `[{"run_id":"r"}]`},
			{"trailing_document", `{"run_id":"r"}{"abandon":true}`},
			{"malformed_json", `{"run_id":"r",`},
			{"body_over_16k", `{"run_id":"` + strings.Repeat("x", 16*1024) + `"}`},
		} {
			t.Run(tc.name, func(t *testing.T) {
				recorder := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest(http.MethodPost, "/api/user/gacha/run/finish", strings.NewReader(tc.payload))
				var intent struct {
					RunID   string `json:"run_id"`
					Abandon bool   `json:"abandon"`
				}
				require.False(t, decodeGachaIntent(c, &intent, []string{"run_id", "abandon"}))
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
				assert.Contains(t, recorder.Body.String(), "GACHA_INVALID_INTENT")
			})
		}
	})

	t.Run("action_quantities_and_team_members_have_exact_types", func(t *testing.T) {
		for _, payload := range []string{
			`{"count":"10"}`, `{"count":1.5}`, `{"count":9223372036854775808}`,
			`{"squad":[{"girl_id":"c01"}]}`, `{"squad":[1]}`, `{"squad":"c01"}`,
		} {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/user/gacha/action", strings.NewReader(payload))
			var intent struct {
				Count *int      `json:"count"`
				Squad *[]string `json:"squad"`
			}
			require.False(t, decodeGachaIntent(c, &intent, []string{"count", "squad"}), payload)
			assert.Equal(t, http.StatusBadRequest, recorder.Code)
		}
	})
}
