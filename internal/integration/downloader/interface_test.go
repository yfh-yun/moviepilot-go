package downloader

import (
	"context"
	"testing"
)

// TestTorrentState 测试种子状态
func TestTorrentState(t *testing.T) {
	tests := []struct {
		state       TorrentState
		downloading bool
		completed   bool
		paused      bool
		error       bool
	}{
		{StateDownloading, true, false, false, false},
		{StateMetaDL, true, false, false, false},
		{StateForcedDL, true, false, false, false},
		{StateUploading, false, true, false, false},
		{StateStalledUP, false, true, false, false},
		{StateCheckingUP, false, true, false, false},
		{StateForcedUP, false, true, false, false},
		{StatePausedDL, false, false, true, false},
		{StatePausedUP, false, false, true, false},
		{StateError, false, false, false, true},
		{StateMissingFiles, false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if got := tt.state.IsDownloading(); got != tt.downloading {
				t.Errorf("IsDownloading() = %v, want %v", got, tt.downloading)
			}
			if got := tt.state.IsCompleted(); got != tt.completed {
				t.Errorf("IsCompleted() = %v, want %v", got, tt.completed)
			}
			if got := tt.state.IsPaused(); got != tt.paused {
				t.Errorf("IsPaused() = %v, want %v", got, tt.paused)
			}
			if got := tt.state.IsError(); got != tt.error {
				t.Errorf("IsError() = %v, want %v", got, tt.error)
			}
		})
	}
}

// TestFactory 测试工厂模式
func TestFactory(t *testing.T) {
	factory := NewFactory()

	// 测试空工厂
	if clients := factory.ListClients(); len(clients) != 0 {
		t.Errorf("新工厂应该为空，但有 %d 个客户端", len(clients))
	}

	// 注册mock客户端
	mock := &mockDownloaderClient{name: "test"}
	factory.Register("test", mock)

	// 测试获取客户端
	client, ok := factory.GetClient("test")
	if !ok {
		t.Error("应该能获取到注册的客户端")
	}
	if client != mock {
		t.Error("获取的客户端不正确")
	}

	// 测试获取不存在的客户端
	_, ok = factory.GetClient("nonexistent")
	if ok {
		t.Error("不应该获取到不存在的客户端")
	}

	// 测试列出客户端
	clients := factory.ListClients()
	if len(clients) != 1 {
		t.Errorf("应该有 1 个客户端，但有 %d 个", len(clients))
	}
	if clients[0] != "test" {
		t.Errorf("客户端名称应该是 'test'，但是 '%s'", clients[0])
	}

	// 注册多个客户端
	factory.Register("qbittorrent", &mockDownloaderClient{name: "qb"})
	factory.Register("transmission", &mockDownloaderClient{name: "tr"})

	clients = factory.ListClients()
	if len(clients) != 3 {
		t.Errorf("应该有 3 个客户端，但有 %d 个", len(clients))
	}
}

// TestFactoryOverwrite 测试覆盖注册
func TestFactoryOverwrite(t *testing.T) {
	factory := NewFactory()

	client1 := &mockDownloaderClient{name: "client1"}
	client2 := &mockDownloaderClient{name: "client2"}

	factory.Register("test", client1)
	factory.Register("test", client2)

	client, ok := factory.GetClient("test")
	if !ok {
		t.Fatal("应该能获取到客户端")
	}

	if mock, ok := client.(*mockDownloaderClient); ok {
		if mock.name != "client2" {
			t.Errorf("应该获取到最后注册的客户端，但得到 %s", mock.name)
		}
	} else {
		t.Error("客户端类型不正确")
	}
}

// mockDownloaderClient 模拟下载器客户端
type mockDownloaderClient struct {
	name string
}

func (m *mockDownloaderClient) AddTorrent(ctx context.Context, req *AddTorrentRequest) (*Torrent, error) {
	return nil, nil
}

func (m *mockDownloaderClient) ListTorrents(ctx context.Context, filter *TorrentFilter) ([]*Torrent, error) {
	return nil, nil
}

func (m *mockDownloaderClient) GetTorrentInfo(ctx context.Context, hash string) (*TorrentInfo, error) {
	return nil, nil
}

func (m *mockDownloaderClient) PauseTorrent(ctx context.Context, hash string) error {
	return nil
}

func (m *mockDownloaderClient) ResumeTorrent(ctx context.Context, hash string) error {
	return nil
}

func (m *mockDownloaderClient) RemoveTorrent(ctx context.Context, hash string, deleteFiles bool) error {
	return nil
}

func (m *mockDownloaderClient) SetTorrentCategory(ctx context.Context, hash string, category string) error {
	return nil
}

func (m *mockDownloaderClient) SetTorrentTags(ctx context.Context, hash string, tags []string) error {
	return nil
}

func (m *mockDownloaderClient) GetVersion(ctx context.Context) (string, error) {
	return "mock-1.0", nil
}

func (m *mockDownloaderClient) TestConnection(ctx context.Context) error {
	return nil
}
