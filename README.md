# AI Test Gap Finder

Ferramenta que analisa alterações em Pull Requests e identifica lacunas de testes no código modificado, comentando no PR (integração GitHub).

## Requisitos

- Go 1.21+

## Variáveis de ambiente

| Variável | Descrição |
|---------|-----------|
| `GITHUB_TOKEN` | Token de acesso ao GitHub (recomendado para evitar rate limit). |
| `GITHUB_WEBHOOK_SECRET` | Secret do webhook para validar assinatura (opcional). |
| `OPENAI_API_KEY` | Chave da API OpenAI para sugestões de cenários (opcional; sem ela usa sugestões heurísticas). |
| `PORT` | Porta HTTP (padrão: 8080). |

## Executar

```bash
go run ./cmd/server
```

Configure o webhook do GitHub App para `POST /webhook`. Eventos: `pull_request` com `action`: `opened` ou `synchronize`.

## Como testar no seu PR

### 1. Subir o servidor e expor na internet

O GitHub precisa conseguir enviar o webhook para a sua máquina. Duas opções:

**Opção A — ngrok (rápido para testar)**

```bash
# Terminal 1: sobe o app
go run ./cmd/server

# Terminal 2: expõe a porta 8080 (instale ngrok se precisar: https://ngrok.com)
ngrok http 8080
```

Anote a URL HTTPS que o ngrok mostrar (ex: `https://abc123.ngrok.io`). A URL do webhook será `https://abc123.ngrok.io/webhook`.

**Opção B — Deploy no Render**  
Veja a seção [Deploy no Render](#deploy-no-render) abaixo.

**Opção C — Outro host**  
Railway, Fly.io, seu servidor, etc. A URL do webhook será `https://seu-dominio.com/webhook`.

### 2. Configurar o webhook no repositório

1. No GitHub: repositório onde quer testar → **Settings** → **Webhooks** → **Add webhook**.
2. **Payload URL:** `https://SUA-URL/webhook` (a URL do ngrok ou do deploy).
3. **Content type:** `application/json`.
4. **Secret (opcional):** gere um valor aleatório e coloque no `.env` como `GITHUB_WEBHOOK_SECRET` (se não usar, deixe em branco).
5. Em **Which events would you like to trigger this webhook?** escolha **Let me select individual events** e marque **Pull requests**.
6. Salve com **Add webhook**.

### 3. Token com permissão para comentar

O `GITHUB_TOKEN` no `.env` precisa ter permissão para **escrever** no repositório (comentar no PR):

- Personal Access Token: scope **repo** (acesso completo aos repos que você escolher).
- O dono do token precisa ter permissão de escrita no repo onde o PR será aberto.

### 4. Disparar o teste

1. Crie um **branch** e altere algum **.go** (por exemplo uma função com `if`).
2. Abra um **Pull Request** para `main` (ou o branch base).
3. O GitHub envia o evento `pull_request` → `opened` para o seu `/webhook`.
4. O servidor analisa o diff e posta um comentário no PR.

Se nada aparecer:

- Em **Settings** → **Webhooks** → clique no webhook e veja **Recent Deliveries**: status 2xx = servidor respondeu; 4xx/5xx ou timeout = conferir servidor e URL.
- No terminal onde o servidor está rodando, veja se há erros de conexão ou da API do GitHub.
- Confirme que o `GITHUB_TOKEN` está no `.env` e que o app carrega o `.env` (você já está usando `godotenv.Load()`).

### Resumo rápido

| Onde | O que fazer |
|------|-------------|
| `.env` | `GITHUB_TOKEN=ghp_...` e, se quiser, `GITHUB_WEBHOOK_SECRET=...` |
| Terminal | `go run ./cmd/server` |
| ngrok | `ngrok http 8080` → copiar URL HTTPS |
| GitHub → Repo → Settings → Webhooks | URL = `https://SUA-URL/webhook`, evento **Pull requests** |
| Repo | Abrir um PR com mudança em `.go` → comentário deve aparecer no PR |

## Deploy no Render (plano gratuito)

Use **Web Service** (não precisa de Blueprint, que é pago).

1. No [Render Dashboard](https://dashboard.render.com/), **New +** → **Web Service**.
2. Conecte o repositório **tonnarruda/ai-test-gap-finder**.
3. O Render detecta Go e cria o serviço. Em **Settings** → **Build & Deploy** altere para:
   - **Build Command:** `go build -ldflags '-s -w' -o server ./cmd/server`
   - **Start Command:** `./server`  
   (O padrão compila a raiz do repo e dá erro; o `main` está em `cmd/server`.)
4. Em **Environment**, adicione:
   - `GITHUB_TOKEN` — seu Personal Access Token com scope **repo**
   - `GITHUB_WEBHOOK_SECRET` — o mesmo valor que você vai usar no webhook do GitHub
   - `OPENAI_API_KEY` — (opcional) chave da OpenAI
5. Salve e faça o **deploy**. Anote a URL (ex: `https://ai-test-gap-finder-xxx.onrender.com`).
6. No GitHub, **Settings** → **Webhooks** do repositório:
   - **Payload URL:** `https://SUA-URL.onrender.com/webhook`
   - **Content type:** `application/json`
   - **Secret:** o mesmo de `GITHUB_WEBHOOK_SECRET`
   - Eventos: **Pull requests**

O Render define `PORT` automaticamente.

## Testes

```bash
go test ./...
```

## Estrutura (PRD)

- `cmd/server` — entrada HTTP e webhook
- `internal/github` — webhook, cliente API (diff, conteúdo, comentário)
- `internal/analyzer` — parser de patch, detecção de funções (AST)
- `internal/testdetector` — localização de testes, detecção de gaps
- `internal/ai` — prompt e motor de sugestões (OpenAI ou mock)
- `internal/commenter` — formatação do comentário no PR
- `internal/app` — pipeline de análise e comentário
