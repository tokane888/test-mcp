# データベース設計

このディレクトリには、モノレポ内の全サービスで共有されるデータベース関連ファイルが含まれています。

## 構造

```
db/
├── init/                  # データベース初期化スクリプト
│   └── 01_create_tables.sql  # 初期スキーマ作成
└── README.md             # このファイル
```

## データベーススキーマ

### Usersテーブル

`users`テーブルは、以下のカラムでユーザーアカウント情報を格納します：

| カラム        | 型                       | 説明                                 |
| ------------- | ------------------------ | ------------------------------------ |
| id            | UUID                     | 主キー、自動生成                     |
| email         | VARCHAR(255)             | ユーザーのメールアドレス（ユニーク） |
| username      | VARCHAR(100)             | ユーザーの表示名                     |
| password_hash | VARCHAR(255)             | Bcryptでハッシュ化されたパスワード   |
| created_at    | TIMESTAMP WITH TIME ZONE | レコード作成時刻                     |
| updated_at    | TIMESTAMP WITH TIME ZONE | 最終更新時刻（自動更新）             |
| deleted_at    | TIMESTAMP WITH TIME ZONE | 論理削除のタイムスタンプ             |

### インデックス

- `id`の主キーインデックス（自動）
- `email`のユニークインデックス（自動）

## 使用方法

### 開発環境

`db/init/`内の初期化スクリプトは、devcontainer環境でPostgreSQLコンテナ起動時に自動的に実行されます。

### 本番環境

#### Docker Composeを使用する場合

```bash
cd db/
cp .env.example .env
# .envファイルを編集してパスワードを設定
docker-compose up -d
```

初期化スクリプトは自動的に実行されます。

#### 既存のPostgreSQLに対して実行する場合

```bash
psql -U postgres -d api_db -f db/init/01_create_tables.sql
```

## タイムゾーンの取り扱い

- データベースは`TIMESTAMP WITH TIME ZONE`を使用して全てのタイムスタンプをUTCで保存
- アプリケーション層で表示用にJSTに変換
- psqlでの表示の利便性のため、PostgreSQLコンテナは`TZ=Asia/Tokyo`で設定
