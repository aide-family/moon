package sms_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/moon-monitor/moon/pkg/plugin"
	"github.com/moon-monitor/moon/pkg/plugin/sms"
)

const mockSmsPlugin = "./mock/mock_sms_plugin.so"

func TestLoadPlugin(t *testing.T) {
	logger := log.NewStdLogger(os.Stdout)

	sender, err := sms.LoadPlugin(&plugin.LoadConfig{
		Path:    mockSmsPlugin,
		Logger:  logger,
		Configs: nil,
	})
	if err != nil {
		t.Fatalf("Failed to load plugin: %v", err)
	}

	msg := sms.Message{
		TemplateParam: "test_param",
		TemplateCode:  "test_code",
	}

	// Test single send
	err = sender.Send(context.Background(), "+1234567890", msg)
	if err != nil {
		t.Errorf("Failed to send message: %v", err)
	}

	// Test batch send
	err = sender.SendBatch(context.Background(), []string{"+123", "+456"}, msg)
	if err != nil {
		t.Errorf("Failed to send batch messages: %v", err)
	}
}

func Test_Async_LoadPlugin_Server(t *testing.T) {
	http.HandleFunc("/load", func(w http.ResponseWriter, r *http.Request) {
		logger := log.NewStdLogger(os.Stdout)

		sender, err := sms.LoadPlugin(&plugin.LoadConfig{
			Path:    mockSmsPlugin,
			Logger:  logger,
			Configs: nil,
		})
		if err != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Failed to load plugin " + err.Error()))
			return
		}

		msg := sms.Message{
			TemplateParam: "test_param",
			TemplateCode:  "test_code",
		}
		if err := sender.Send(context.Background(), "+1234567890", msg); err != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Failed to send message " + err.Error()))
			return
		}
	})
	http.HandleFunc("/other", func(w http.ResponseWriter, r *http.Request) {
		filename := fmt.Sprintf("%s.so", r.URL.Query().Get("name"))
		// 判断文件是否存在， 存在则删除
		if _, err := os.Stat(filename); err == nil {
			if err := os.Remove(filename); err != nil {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Failed to remove file " + err.Error()))
				return
			}
		}
		// 从http里获取 .so文件流， 并保存到当前目录
		file, err := os.Create(filename)
		if err != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Failed to create file " + err.Error()))
			return
		}
		defer file.Close()

		_, err = io.Copy(file, r.Body)
		if err != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Failed to copy file " + err.Error()))
			return
		}

		logger := log.NewStdLogger(os.Stdout)

		sender, err := sms.LoadPlugin(&plugin.LoadConfig{
			Path:    filename,
			Logger:  logger,
			Configs: nil,
		})
		if err != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Failed to load plugin " + err.Error()))
			return
		}

		msg := sms.Message{
			TemplateParam: "test_param",
			TemplateCode:  "test_code",
		}
		if err := sender.Send(context.Background(), "+1234567890", msg); err != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Failed to send message " + err.Error()))
			return
		}
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		t.Errorf("Failed to start server: %v", err)
		return
	}
}

func Test_Async_LoadPlugin(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.NewStdLogger(os.Stdout)

		sender, err := sms.LoadPlugin(&plugin.LoadConfig{
			Path:    mockSmsPlugin,
			Logger:  logger,
			Configs: nil,
		})
		if err != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Failed to load plugin " + err.Error()))
			return
		}

		msg := sms.Message{
			TemplateParam: "test_param",
			TemplateCode:  "test_code",
		}
		if err := sender.Send(context.Background(), "+1234567890", msg); err != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Failed to send message " + err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}))
	defer server.Close()

	// Make a request to the test server
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Verify the response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if string(body) != "Success" {
		t.Errorf("Unexpected response body: %s", string(body))
	}
}
