# kubecloud

KubeCloud is a CLI tool that helps you deploy and manage Kubernetes clusters on the decentralized TFGrid.

## Requirements

- Go >= 1.21
- docker

## Configuration

Before building or running backend create `config.json` in `backend` dir.

example `config.json`:

```json
{
  "server": {
    "host": "localhost", //should be changed if run using docker-compose to be `0.0.0.0`
    "port": "8080"
  },
  "database": {
    "file": "<the path of the database file you have or you want to create, default is `users.db`>"
  },
  "token": {
    "secret": "<your secret for the jwt tokens, required>",
    "access_token_expiry_minutes": "<timeout for access token in minutes>",
    "refresh_token_expiry_hours": "<timeout for refresh token in minutes>"
  },
  "admins": ["<a set of the user emails you want to make admins>"],
  "mailSender": {
    "email": "your sendgrid account sender, required",
    "sendgrid_key":  "<sendgrid-key>, required",
    "timeout": "<the timeout for app mail verification codes in seconds, required>"
  },
  "currency": "<currency users will paid with in stripe>",
  "stripe_secret": "<stripe secret for account>",
  "tfchain_url": "wss://tfchain.dev.grid.tf/wss",
  "gridproxy_url": "https://gridproxy.dev.grid.tf/",
  "voucher_name_length": 5,
  "terms_and_conditions": {
    "document_link": "https://manual.grid.tf/labs/knowledge_base/terms_conditions_all3",
    "document_hash": "6f2b4109704ba2883d978a7b94e5f295"
  },
  "activation_service_url": "https://activation.dev.grid.tf/activation/activate",
  "system_account": {
    "mnemonics": "<system account mnemonic>",
    "network": "dev"
  },
  "graphql_url": "https://graphql.dev.grid.tf/graphql",
  "firesquid_url": "https://firesquid.dev.grid.tf/graphql",
  "redis": {
    "host": "localhost", // Redis host should be changed if running using docker-compose to be `redis`
    "port": 6379,
    "password": "",
    "db": 0
  },
  "deployer_workers_num": 3
}
```

### Run using docker 

- Navigate to root dir

```bash
docker-compose up
```