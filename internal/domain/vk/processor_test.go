package vk

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/boliev/graphai/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestProcessor_sendWithRetries(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success on first try", func(t *testing.T) {
		t.Parallel()

		aiMock := newMockai(t)
		p := &Processor{
			aiClient: aiMock,
			logger:   logger,
		}

		prompt := "make photo better"
		files := []string{"file1.jpg", "file2.jpg"}
		expected := &domain.AIResponse{
			Photo: []byte("result"),
		}

		aiMock.EXPECT().
			Send(ctx, prompt, files).
			Return(expected, nil).
			Once()

		got, err := p.sendWithRetries(ctx, prompt, files)
		require.NoError(t, err)
		require.Equal(t, expected, got)
	})

	t.Run("success on third try", func(t *testing.T) {
		t.Parallel()

		aiMock := newMockai(t)
		p := &Processor{
			aiClient: aiMock,
			logger:   logger,
		}

		prompt := "make photo better"
		files := []string{"file1.jpg"}
		expected := &domain.AIResponse{
			Photo: []byte("result"),
		}
		sendErr := errors.New("temporary ai error")

		aiMock.EXPECT().
			Send(ctx, prompt, files).
			Return(nil, sendErr).
			Once()

		aiMock.EXPECT().
			Send(ctx, prompt, files).
			Return(nil, sendErr).
			Once()

		aiMock.EXPECT().
			Send(ctx, prompt, files).
			Return(expected, nil).
			Once()

		got, err := p.sendWithRetries(ctx, prompt, files)
		require.NoError(t, err)
		require.Equal(t, expected, got)
	})

	t.Run("all retries failed", func(t *testing.T) {
		t.Parallel()

		aiMock := newMockai(t)
		p := &Processor{
			aiClient: aiMock,
			logger:   logger,
		}

		prompt := "make photo better"
		files := []string{"file1.jpg"}
		sendErr := errors.New("ai failed")

		aiMock.EXPECT().
			Send(ctx, prompt, files).
			Return(nil, sendErr).
			Times(geminiResentTries)

		got, err := p.sendWithRetries(ctx, prompt, files)
		require.Error(t, err)
		require.ErrorIs(t, err, sendErr)
		require.Nil(t, got)
	})

	t.Run("returns nil nil when ai returns nil nil", func(t *testing.T) {
		t.Parallel()

		aiMock := newMockai(t)
		p := &Processor{
			aiClient: aiMock,
			logger:   logger,
		}

		prompt := "make photo better"
		files := []string{"file1.jpg"}

		aiMock.EXPECT().
			Send(ctx, prompt, files).
			Return(nil, nil).
			Once()

		got, err := p.sendWithRetries(ctx, prompt, files)
		require.NoError(t, err)
		require.Nil(t, got)
	})
}

func TestProcessor_userFromMessage(t *testing.T) {
	t.Parallel()

	p := &Processor{}

	msg := object.MessagesMessage{
		FromID: 12345,
		PeerID: 67890,
	}

	got := p.userFromMessage(msg)

	require.NotNil(t, got)
	require.Equal(t, int64(12345), got.UserVKID)
	require.Equal(t, int64(67890), got.PeerID)
}

func TestProcessor_command(t *testing.T) {
	t.Parallel()

	p := &Processor{}

	t.Run("invalid payload", func(t *testing.T) {
		t.Parallel()

		msg := object.MessagesMessage{
			Payload: `{"cmd":`,
		}

		err := p.command(msg)
		require.Error(t, err)
		require.ErrorContains(t, err, "invalid payload")
		require.ErrorContains(t, err, `raw={"cmd":`)
	})

	t.Run("empty payload becomes help branch and panics without sender", func(t *testing.T) {
		t.Parallel()

		msg := object.MessagesMessage{
			Payload: `{}`,
		}

		require.Panics(t, func() {
			_ = p.command(msg)
		})
	})
}
