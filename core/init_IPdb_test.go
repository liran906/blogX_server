// core/init_ipdb_test.go
package core

import (
	"os"
	"testing"
)

func TestGetAddress(t *testing.T) {
	_ = os.Chdir("..")
	currentDir, _ := os.Getwd()
	t.Logf("当前工作目录: %s", currentDir)

	InitIPDB()

	tests := []struct {
		name    string
		ip      string
		want    string
		wantErr bool
	}{

		{
			name:    "无效IP格式 1",
			ip:      "invalid.ip",
			want:    "",
			wantErr: true,
		},
		{
			name:    "无效IP格式 2",
			ip:      "0.0.0.",
			want:    "",
			wantErr: true,
		},
		{
			name:    "无效IP格式 3",
			ip:      "256.123.123.123",
			want:    "",
			wantErr: true,
		},
		{
			name:    "无效IP格式 4",
			ip:      "-1.2.3.4",
			want:    "",
			wantErr: true,
		},
		{
			name:    "内网IP 1",
			ip:      "192.168.1.1",
			want:    "Private IP",
			wantErr: false,
		},
		{
			name:    "内网IP 2",
			ip:      "0.0.0.0",
			want:    "Private IP",
			wantErr: false,
		},
		{
			name:    "内网IP 3",
			ip:      "127.0.0.1",
			want:    "Private IP",
			wantErr: false,
		},
		{
			name:    "国内IP 1",
			ip:      "123.123.123.123",
			want:    "中国-北京",
			wantErr: false,
		},
		{
			name:    "国内IP 2",
			ip:      "114.114.114.114",
			want:    "江苏-南京",
			wantErr: false,
		},
		{
			name:    "国内IP 3",
			ip:      "101.226.168.228",
			want:    "中国-上海",
			wantErr: false,
		},
		{
			name:    "国内IP 4",
			ip:      "182.239.127.137",
			want:    "中国-香港",
			wantErr: false,
		},
		{
			name:    "国内IP  5",
			ip:      "153.3.126.142",
			want:    "江苏-南京",
			wantErr: false,
		},
		{
			name:    "国内IP  6",
			ip:      "134.208.0.0",
			want:    "中国-台湾",
			wantErr: false,
		},

		{
			name:    "国外IP 1",
			ip:      "104.104.51.255",
			want:    "乌拉圭",
			wantErr: false,
		},
		{
			name:    "国外IP 2",
			ip:      "12.44.51.255",
			want:    "美国-新泽西",
			wantErr: false,
		},
		{
			name:    "国外IP 3",
			ip:      "220.44.51.255",
			want:    "日本-兵库",
			wantErr: false,
		},
		{
			name:    "国外IP 4",
			ip:      "188.44.51.255",
			want:    "俄罗斯-莫斯科",
			wantErr: false,
		},
		{
			name:    "国外IP 5",
			ip:      "137.132.213.137",
			want:    "新加坡",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLocationFromIP(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLocationFromIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetLocationFromIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
