# PRD — AI Test Gap Finder

## 1. Visão do Produto

O **AI Test Gap Finder** é uma ferramenta que analisa automaticamente alterações em Pull Requests e identifica **lacunas de testes (test gaps)** no código modificado.

A ferramenta comenta diretamente no PR apontando:

- cenários não testados
- branches não cobertas
- possíveis edge cases
- sugestões de novos testes

Integração inicial com GitHub.

**Objetivo:** ajudar desenvolvedores e QAs a evitar regressões antes do merge.

---

# 2. Problema

Ferramentas atuais mostram apenas **métricas de cobertura**, mas não dizem:

- quais cenários não estão testados
- quais caminhos lógicos estão sem teste
- quais testes deveriam existir

Resultado:

- cobertura alta com testes fracos
- bugs escapando para produção

---

# 3. Objetivo do MVP

Detectar **lacunas de testes nas alterações de um Pull Request** e postar um comentário automático no PR.

Escopo inicial:

- analisar código alterado
- identificar branches
- verificar presença de testes
- sugerir cenários faltantes

---

# 4. Fluxo do Usuário

## Passo 1
Usuário instala o GitHub App no repositório.

## Passo 2
Usuário abre um Pull Request.

## Passo 3
Sistema executa análise automática.

## Passo 4
Ferramenta comenta no PR.

Exemplo:


AI Test Gap Finder Report

Files analyzed: 3

⚠ Potential Test Gaps

File: user_service.go

Function: ValidateLogin

Missing scenarios:

invalid password

empty password

expired token

user not found

Suggested tests:

TestValidateLogin_InvalidPassword
TestValidateLogin_UserNotFound


---

# 5. Requisitos Funcionais

## RF01 — Detectar Pull Requests

Sistema deve escutar eventos de PR via webhook do GitHub.

Eventos:

- pull_request opened
- pull_request synchronize

---

## RF02 — Obter diff do PR

Sistema deve obter:

- arquivos modificados
- linhas alteradas
- contexto da função

---

## RF03 — Analisar funções alteradas

Sistema identifica:

- funções modificadas
- condições if
- branches

Exemplo:

```go
if user == nil
if password == ""
if tokenExpired
RF04 — Verificar existência de testes

Sistema busca:

arquivos *_test.go

funções de teste relacionadas

RF05 — Detectar lacunas

Heurísticas iniciais:

branch sem teste

condição sem cenário

erro não testado

RF06 — Gerar sugestões

Exemplo:

Missing scenarios:

password empty
password invalid
user locked
RF07 — Postar comentário no PR

Sistema cria comentário via API do GitHub.

6. Requisitos Não Funcionais
Performance

Análise deve levar:

< 30 segundos

por PR.

Segurança

acesso somente leitura ao repositório

tokens seguros

Escalabilidade

Arquitetura preparada para:

100 PR analyses / hour
7. Arquitetura Inicial
GitHub Webhook
      ↓
API Server (Go)
      ↓
PR Analyzer
      ↓
Test Gap Detector
      ↓
AI Suggestion Engine
      ↓
PR Comment Publisher
8. Stack Tecnológica

Backend:

Go

Integração:

GitHub REST API

IA:

OpenAI API

Parser de código:

Go AST
9. Estrutura de Projeto (Go)
ai-test-gap-finder
│
├── cmd
│   └── server
│
├── internal
│
│   ├── github
│   │   ├── webhook_handler.go
│   │   ├── pr_client.go
│
│   ├── analyzer
│   │   ├── diff_parser.go
│   │   ├── function_detector.go
│
│   ├── testdetector
│   │   ├── test_finder.go
│   │   ├── gap_detector.go
│
│   ├── ai
│   │   ├── prompt_builder.go
│   │   ├── suggestion_engine.go
│
│   └── commenter
│       └── github_comment.go
10. Exemplo de Prompt da IA

Entrada:

func ValidateLogin(user *User, password string) error {
   if user == nil {
      return ErrUserNotFound
   }

   if password == "" {
      return ErrInvalidPassword
   }

   return nil
}

Prompt:

Identify missing test scenarios for this function.
Return a list of test cases that should exist.
11. Métricas de Sucesso

Primeira versão:

10 repos usando

Depois:

100 PR analisados por semana
12. Roadmap Futuro
v1

PR comments

v2

dashboard de risco

v3

geração automática de testes

v4

integração com GitLab e Bitbucket

13. Exemplo de Comentário Final no PR
AI Test Gap Finder

Analysis Summary

Files changed: 4
Functions analyzed: 7

Potential Missing Tests:

UserService.ValidateLogin

- empty password
- user not found
- expired session

Recommendation:

Add test cases for these scenarios to improve coverage and reduce regression risk.

---

💡 Se quiser, eu também posso te entregar o **PRD v2 mais profissional**, com:

- **User Stories**
- **Critérios de Aceite**
- **Modelo de domínio**
- **Algoritmo de detecção de gaps**
- **Event flow do GitHub App**

Esse já fica **nível produto pronto para crescer** e ajuda muito na hora de implementar.