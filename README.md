# test-mcp

試験的に導入したMCPの導入手順などを記載

## chrome-mcp

- chromeを操作するMCP
- devcontainer上ではchromeの待受側portへアクセスできないため使用不可

### 導入手順

- MCP server側
  - [mcp-chrome拡張](https://github.com/hangwin/mcp-chrome?tab=readme-ov-file)をQuick Startの手順にしたがってインストール
  - `~/.config/google-chrome/Default/Extensions/`配下
    - 通常はここにchrome拡張の本体があるが、今回は何もない
      - 今回の手順ではパッケージ化されていないchrome拡張をinstallしているため
  - chrome拡張にnative messaging host関連の権限が付与されていることを確認
    - 例

      ```sh
      ~/.config/google-chrome/NativeMessagingHosts cat com.chromemcp.nativehost.json
      {
        "name": "com.chromemcp.nativehost",
        "description": "Node.js Host for Browser Bridge Extension",
        "path": "/usr/local/lib/node_modules/mcp-chrome-bridge/dist/run_host.sh",
        "type": "stdio",
        "allowed_origins": [
          "chrome-extension://hbdgbgagpkpjffpklnamcljpakneikee/"
        ]
      }%
      ```

    - chrome拡張から呼び出すrun_host.shのownerをchromeのownerと同一に設定
      - `sudo chown tom:tom /usr/local/lib/node_modules/mcp-chrome-bridge/dist/run_host.sh`
        - `tom:tom`としている部分は実際のownerに置き換え
- MCP client側
  - mcp-chrome-bridgeインストール
    - `sudo npm install -g mcp-chrome-bridge`
    - ログ出力先ディレクトリ作成
      - `sudo mkdir -p /usr/local/lib/node_modules/mcp-chrome-bridge/dist/logs`
      - `sudo chmod 777 /usr/local/lib/node_modules/mcp-chrome-bridge/dist/logs`
  - claude codeにMCP追加
    - `claude mcp add chrome node /usr/local/lib/node_modules/mcp-chrome-bridge/dist/mcp/mcp-server-stdio.js`
    - MCPへの接続成功確認
      - 例

        ```sh
         ~ claude mcp list
        Checking MCP server health...

        chrome: node /usr/local/lib/node_modules/mcp-chrome-bridge/dist/mcp/mcp-server-stdio.js - ✓ Connected
        ```
