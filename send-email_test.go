package main

import (
	"context"
	"encoding/json"
	"html/template"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hivemindd/communication-service/config"
	"github.com/hivemindd/communication-service/internal/form"
	"github.com/hivemindd/communication-service/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// Helper to create a dummy logger
func dummyLogger(t *testing.T) *zap.SugaredLogger {
	return zaptest.NewLogger(t).Sugar()
}

// Helper to create a dummy template
func dummyTemplate() *template.Template {
	return template.Must(template.New("test").Parse("Hello"))
}

// Helper to create a dummy config
func dummyConfig() *config.Config {
	return &config.Config{
		UIBaseURL: "http://localhost",
	}
}

// Helper to create a dummy amqp091.Delivery with testify spies
type testDelivery struct {
	acked    bool
	nacked   bool
	rejected bool
	body     []byte
}

func (d *testDelivery) Ack(multiple bool) error           { d.acked = true; return nil }
func (d *testDelivery) Nack(multiple, requeue bool) error { d.nacked = true; return nil }
func (d *testDelivery) Reject(requeue bool) error         { d.rejected = true; return nil }
func (d *testDelivery) GetBody() []byte                   { return d.body }

func TestHandleMsg_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockEmailProvider(ctrl)
	logger := dummyLogger(t)
	conf := dummyConfig()
	tmpl := dummyTemplate()

	email := form.Email{
		Type: "forgot_password",
		To:   "test@example.com",
	}
	body, _ := json.Marshal(email)
	d := &testDelivery{body: body}

	templateMap := map[string]*template.Template{
		"forgot_password": tmpl,
	}

	mockProvider.EXPECT().Send(gomock.Any(), tmpl).Return(nil)

	handleMsg(context.Background(), conf, d, logger, mockProvider, templateMap)

	require.True(t, d.acked, "Delivery should be acked on success")
	require.False(t, d.nacked, "Delivery should not be nacked on success")
	require.False(t, d.rejected, "Delivery should not be rejected on success")
}

func TestHandleMsg_SendError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockEmailProvider(ctrl)
	logger := dummyLogger(t)
	conf := dummyConfig()
	tmpl := dummyTemplate()

	email := form.Email{
		Type: "forgot_password",
		To:   "test@example.com",
	}
	body, _ := json.Marshal(email)
	d := &testDelivery{body: body}

	templateMap := map[string]*template.Template{
		"forgot_password": tmpl,
	}

	mockProvider.EXPECT().Send(gomock.Any(), tmpl).Return(assert.AnError)

	handleMsg(context.Background(), conf, d, logger, mockProvider, templateMap)

	require.False(t, d.acked, "Delivery should not be acked on error")
	require.True(t, d.nacked, "Delivery should be nacked on error")
	require.False(t, d.rejected, "Delivery should not be rejected on error")
}

func TestHandleMsg_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockEmailProvider(ctrl)
	logger := dummyLogger(t)
	conf := dummyConfig()

	d := &testDelivery{body: []byte("not-json")}
	templateMap := map[string]*template.Template{}

	handleMsg(context.Background(), conf, d, logger, mockProvider, templateMap)

	require.False(t, d.acked, "Delivery should not be acked on invalid JSON")
	require.False(t, d.nacked, "Delivery should not be nacked on invalid JSON")
	require.True(t, d.rejected, "Delivery should be rejected on invalid JSON")
}

func TestHandleMsg_MissingTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockEmailProvider(ctrl)
	logger := dummyLogger(t)
	conf := dummyConfig()

	email := form.Email{
		Type: "unknown_type",
		To:   "test@example.com",
	}
	body, _ := json.Marshal(email)
	d := &testDelivery{body: body}

	templateMap := map[string]*template.Template{}

	handleMsg(context.Background(), conf, d, logger, mockProvider, templateMap)

	require.False(t, d.acked, "Delivery should not be acked on missing template")
	require.False(t, d.nacked, "Delivery should not be nacked on missing template")
	require.True(t, d.rejected, "Delivery should be rejected on missing template")
}

func TestHandleMsg_TemplateFoundButNotInMap(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockEmailProvider(ctrl)
	logger := dummyLogger(t)
	conf := dummyConfig()

	email := form.Email{
		Type: "forgot_password",
		To:   "test@example.com",
	}
	body, _ := json.Marshal(email)
	d := &testDelivery{body: body}

	// templateMap does NOT contain "forgot_password"
	templateMap := map[string]*template.Template{}

	// Expect Send to be called with nil template
	mockProvider.EXPECT().Send(gomock.Any(), nil).Return(nil)

	handleMsg(context.Background(), conf, d, logger, mockProvider, templateMap)

	require.True(t, d.acked, "Delivery should be acked even if template is nil in map")
	require.False(t, d.nacked, "Delivery should not be nacked")
	require.False(t, d.rejected, "Delivery should not be rejected")
}
