# Configuração, execução e testes

[Índice da documentação](README.md) · [Arquitetura](arquitetura.md)

## Requisitos e preparação

- Toolchain Go compatível com a versão declarada no [go.mod](../go.mod)
  (atualmente `1.26.1`).
- PostgreSQL com a extensão pgvector instalada e habilitada no banco da API.
- Credenciais Google e segredos locais para JWT e sessão OAuth.
- Ollama e o modelo `mxbai-embed-large` para gerar recomendações; a ausência do
  Ollama não impede o cadastro básico de demandas.

Execute os comandos a partir da raiz do repositório. As variáveis podem vir do
ambiente do processo ou do arquivo `.env`, carregado por `godotenv/autoload`.
O arquivo `.env` é local e ignorado pelo Git. Os exemplos abaixo não contêm
credenciais reais.

## Variáveis

| Variável | Uso |
| --- | --- |
| `PORT` | Porta HTTP. Configure explicitamente, por exemplo `3333`; não há fallback fixo no código |
| `LOCAL_DB` | `true` monta a conexão com `BLUEPRINT_DB_*`; ausente ou `false` usa `DATABASE_URL` |
| `BLUEPRINT_DB_HOST` | Host PostgreSQL quando `LOCAL_DB=true` |
| `BLUEPRINT_DB_PORT` | Porta PostgreSQL; também usada no mapeamento de porta do Compose |
| `BLUEPRINT_DB_DATABASE` | Nome do banco local |
| `BLUEPRINT_DB_USERNAME` | Usuário do banco local |
| `BLUEPRINT_DB_PASSWORD` | Senha do banco local |
| `BLUEPRINT_DB_SCHEMA` | Opcional; define `search_path` na conexão local, mas não cria o schema |
| `DATABASE_URL` | URL PostgreSQL obrigatória quando o modo local não está habilitado; inclui as opções de TLS desejadas |
| `JWT_SECRET_KEY` | Segredo de assinatura, com pelo menos 32 caracteres após remoção dos espaços nas extremidades |
| `JWT_SECRECT_KEY` | Nome antigo, aceito apenas quando `JWT_SECRET_KEY` não está definida |
| `GOOGLE_CLIENT_ID` | Audiência aceita para o ID token e client ID do OAuth manual |
| `GOOGLE_CLIENT_SECRET` | Credencial do OAuth manual; a inicialização atual exige seu preenchimento |
| `SESSION_SECRET` | Segredo de ao menos 32 caracteres para cookies do OAuth manual |
| `APP_ENV` | Quando igual a `production`, marca o cookie OAuth como `Secure` |
| `OLLAMA_URL` | Endereço do Ollama, por exemplo `http://localhost:11434` |
| `TEST_DATABASE_URL` | Conexão exclusiva dos testes de integração; não substitui a conexão da API |

Exemplo de estrutura para configurar o ambiente local, substituindo todos os
marcadores antes de executar:

```dotenv
PORT=3333
LOCAL_DB=true
BLUEPRINT_DB_HOST=localhost
BLUEPRINT_DB_PORT=5432
BLUEPRINT_DB_DATABASE=nexus
BLUEPRINT_DB_USERNAME=nexus
BLUEPRINT_DB_PASSWORD=<senha-local>
BLUEPRINT_DB_SCHEMA=public
JWT_SECRET_KEY=<segredo-aleatorio-com-32-ou-mais-caracteres>
GOOGLE_CLIENT_ID=<client-id-google>
GOOGLE_CLIENT_SECRET=<client-secret-google>
SESSION_SECRET=<outro-segredo-aleatorio-com-32-ou-mais-caracteres>
APP_ENV=development
OLLAMA_URL=http://localhost:11434
```

No modo `LOCAL_DB=true`, o código usa `sslmode=disable`. Para outra configuração
de conexão, use `LOCAL_DB=false` e uma `DATABASE_URL` adequada ao ambiente.

## Banco e Docker

O [docker-compose.yml](../docker-compose.yml) inicia apenas PostgreSQL; não inicia
a API nem Ollama. A imagem atual é `postgres:latest`, e o Compose não instala nem
habilita pgvector. Portanto, esse arquivo isoladamente ainda não prepara todos
os requisitos da API.

Depois que pgvector estiver instalado no servidor PostgreSQL, execute no banco
que a API utilizará:

```sql
CREATE EXTENSION IF NOT EXISTS vector;
SELECT extversion FROM pg_extension WHERE extname = 'vector';
```

Se a extensão não estiver disponível na instalação, o primeiro comando falhará;
é necessário preparar um PostgreSQL com pgvector antes de executar as migrações.
Mesmo sem usar recomendações, o esquema atual possui colunas `vector(1024)`.

Comandos existentes do Compose:

```sh
docker compose up -d
docker compose down
```

O banco usa o volume nomeado `psql_volume_bp`. As alterações de esquema são
executadas na inicialização e no seed; não há comando de rollback versionado.

## Google e login manual

O aplicativo deve obter um ID token cuja audiência corresponda a
`GOOGLE_CLIENT_ID`, conforme [autenticacao.md](autenticacao.md).

Para testar pelo navegador, abra `http://localhost:3333/login`. O callback atual
é fixo em `http://localhost:3333/auth/google/callback`, definido em
`internal/auth/auth.go`; essa URL deve estar autorizada na configuração OAuth.
Alterar somente `PORT` não altera o callback.

O callback retorna o ID token Google. Ele ainda precisa ser enviado em
`POST /login/v2` para obter o JWT da API. `SESSION_SECRET` protege o cookie desse
fluxo no navegador; `JWT_SECRET_KEY` assina os tokens usados pelo aplicativo.

## Executar os comandos

```sh
go mod download
go run ./cmd/api
```

Para compilar sem iniciar:

```sh
go build ./cmd/api
```

Dados de demonstração e reprocessamento de embeddings são comandos separados:

```sh
go run ./cmd/seed
go run ./cmd/embeddings
```

O seed cria exemplos de áreas, contas, perfis e demandas. Ele não fornece um
login autenticado para essas contas nem associa automaticamente identidades
Google. O reprocessamento tenta gerar os vetores pendentes; não é um daemon.

| Atalho Make | Equivalente/finalidade |
| --- | --- |
| `make run` | Executar a API |
| `make build` | Compilar para `main` |
| `make test` | `go test ./... -v` |
| `make all` | Compilar e testar |
| `make itest` | Testes de domínio cujo nome contém `Postgres` |
| `make seed` | Dados de demonstração e esquema |
| `make embeddings` | Reprocessar vetores |
| `make docker-run` / `make docker-down` | Subir/parar o banco do Compose |
| `make watch` | Executar Air usando `.air.toml` |
| `make clean` | Remover o binário `main` |

Alguns alvos Make usam comandos de shell Unix. No PowerShell, os comandos Go e
Docker diretos acima evitam depender desse shell.

## Testes

```sh
go test ./...
```

Sem `TEST_DATABASE_URL`, os testes de integração com PostgreSQL são ignorados;
um resultado geral de sucesso não confirma que eles tenham sido executados.
Para executá-los em PowerShell:

```powershell
$env:TEST_DATABASE_URL = 'postgres://usuario:senha@localhost:5432/nexus_test?sslmode=disable'
go test ./internal/domain/... -run Postgres -count=1 -v
```

Use um banco dedicado de testes. O teste de recomendações requer pgvector; o de
demandas cria um schema isolado e exige permissão para criá-lo. Os testes de
identidade e perfis usam tabelas temporárias. Os testes de demandas não validam
a migração do esquema real. Os testes unitários usam substitutos para serviços
externos e não comprovam login Google ou Ollama reais.

## Diagnóstico

| Sintoma | Verificação |
| --- | --- |
| Falha de inicialização por JWT/OAuth | Conferir nomes, presença e tamanho dos segredos e credenciais |
| Erro `vector` inexistente | Conferir instalação e habilitação de pgvector no banco correto |
| `/health` informa `down` | Conferir conexão, host, porta, credenciais e disponibilidade do PostgreSQL |
| Login retorna `409` | Conta existente sem vínculo Google correspondente; não há associação automática por e-mail |
| Operação retorna `403` | Consultar `/me` e conferir papel, proprietário, perfil e bloqueio |
| Ação de demanda retorna `409` | Atualizar a demanda e verificar estado, atribuição e prazo da análise |
| Recomendações vazias | Conferir áreas do professor, demandas elegíveis e vetores pendentes |
| Navegador bloqueia por CORS | A origem atualmente autorizada é apenas `http://localhost:5173` |

`GET /health` sempre responde HTTP `200` no handler atual, mesmo quando o corpo
contém `status: "down"`. Para esse diagnóstico, examine o campo `status`, não
apenas o código HTTP. O servidor da aplicação usa HTTP; TLS e implantação externa
não são configurados por este repositório.
