# Referência de rotas

[Índice da documentação](README.md) · [Arquitetura](arquitetura.md)

Os caminhos são relativos ao endereço da API, por exemplo
`http://localhost:3333`. `:id` representa um ID numérico positivo. Use as barras
finais mostradas nas rotas de coleção para evitar redirecionamentos do Gin.

Entradas JSON devem usar `Content-Type: application/json`. Exceto nas rotas
públicas indicadas abaixo, envie `Authorization: Bearer <access_token>`.
O token é o JWT da API, não o ID token Google usado no login.

## Públicas e sessão

| Método | Rota | Acesso | Resultado de sucesso |
| --- | --- | --- | --- |
| GET | `/` | Público | `200`, `{"message":"Hello World"}`; endpoint demonstrativo |
| GET | `/health` | Público | `200`, diagnóstico de conexão e pool do banco |
| GET | `/login` | Público | `200`, página HTML para o login Google manual |
| GET | `/auth/google` | Público | Inicia o redirecionamento OAuth Google |
| GET | `/auth/google/callback` | Público/OAuth | `200`, `id_token` Google e `provider`; falha retorna `401` |
| POST | `/login/v2` | ID token Google no corpo | `200`, JWT, validade e dados do usuário |
| GET | `/me` | JWT válido | `200`, `{"user":{...}}`, inclusive para conta bloqueada |

Login do aplicativo:

```http
POST /login/v2
Content-Type: application/json

{"idToken":"ID_TOKEN_GOOGLE"}
```

O JWT retornado vale 12 horas. Não há endpoints de refresh ou logout; após a
expiração, faça novo login. Logout do app remove a cópia local do token.
Contrato completo e códigos de erro: [autenticacao.md](autenticacao.md).

O callback manual entrega uma identidade Google, não uma sessão da API. A falha
desse callback usa `{"error":"Google authentication failed"}`, diferente do
envelope de erro dos endpoints principais.

## Perfis de demandantes e professores

| Método | Rota | Acesso | Resultado de sucesso |
| --- | --- | --- | --- |
| POST | `/applicants/` | Demandante autenticado | `201`, `{"applicantID":1}` |
| GET | `/applicants/` | Administrador | `200`, lista de DTOs de demandantes |
| GET | `/applicants/:id` | Próprio usuário ou administrador | `200`, DTO do demandante |
| PUT | `/applicants/:id` | Demandante proprietário | `204`, sem corpo |
| POST | `/teachers/` | Professor autenticado | `201`, `{"teacherID":2}` |
| GET | `/teachers/` | Administrador | `200`, lista de DTOs de professores |
| GET | `/teachers/:id` | Próprio usuário ou administrador | `200`, DTO do professor |
| PUT | `/teachers/:id` | Professor proprietário | `204`, sem corpo |

O POST completa o perfil da conta autenticada, sem receber ID de outra pessoa.
O PUT recebe o formulário completo dos campos editáveis. Não há DELETE de
perfil nem permissão administrativa de edição nesses endpoints.

Os DTOs incluem nome, e-mail, papel e campos próprios de cada perfil; seu acesso
é restrito. Exemplos de entrada, critérios de preenchimento e regras de
foto/SIAPE/tipos de demandante estão em [perfis.md](perfis.md).

## Áreas

| Método | Rota | Acesso atual | Resultado de sucesso |
| --- | --- | --- | --- |
| GET | `/areas/` | Usuário autenticado não bloqueado | `200`, lista de áreas |
| GET | `/areas/:id` | Usuário autenticado não bloqueado | `200`, área |
| POST | `/areas/` | Usuário autenticado não bloqueado | `201`, `{"areaID":1}` |
| PUT | `/areas/:id` | Usuário autenticado não bloqueado | `204`, sem corpo |
| DELETE | `/areas/:id` | Usuário autenticado não bloqueado | `204`, sem corpo |

Corpo básico de criação/edição:

```json
{"name":"Desenvolvimento de Software"}
```

Resposta de consulta de uma área sem vetor:

```json
{"id":1,"name":"Desenvolvimento de Software","Embedding":null}
```

O CRUD atual não restringe mutações de áreas ao administrador, não normaliza
nomes nem garante unicidade de nomes. Ele recebe o modelo diretamente, sem a
decodificação estrita dos DTOs de perfis/demandas. Envie apenas o campo de
conteúdo `name`; o vetor é gerado pelo servidor.

A leitura ainda expõe `Embedding`, com essa capitalização. Lista vazia pode ser
`null`. Exclusão é física e pode falhar por vínculos no banco. Erros de ausência
ou de banco ainda podem retornar `500` com detalhes internos; não há a mesma
padronização de demandas nesse módulo.

## Áreas do professor

| Método | Rota | Acesso | Resultado de sucesso |
| --- | --- | --- | --- |
| GET | `/areas-teachers/professor/:id` | Próprio usuário ou administrador | `200`, lista de associações |
| PUT | `/areas-teachers/professor/:id` | Professor proprietário | `204`, sem corpo |
| PUT | `/teachers/:id/areas` | Professor proprietário | `204`, alias da substituição de associações |

Corpo dos dois PUTs:

```json
{
  "areas": [
    {"areas_id":1,"type":"técnica"},
    {"areas_id":2,"type":"interesse"}
  ]
}
```

A operação substitui toda a lista de forma transacional. IDs devem existir e
não podem se repetir. Tipos aceitos: `técnica` e `interesse`, com normalização de
espaços e caixa. `{"areas":[]}` remove todas as associações; campo ausente ou
`null` é inválido. Essa rota específica não exige uma área de cada tipo, ao
contrário do formulário completo de perfil.

Cada item da consulta contém `areas_id`, `name` e `type`; lista vazia é `[]`.
Entradas inválidas retornam `400`; professor ou área inexistente na substituição
retornam `404`. A consulta de associações, isoladamente, não distingue professor
inexistente de professor sem associações.

## Demandas

| Método | Rota | Acesso/regra | Resultado de sucesso |
| --- | --- | --- | --- |
| POST | `/demands/` | Demandante com perfil | `201`, DTO da demanda |
| GET | `/demands/` | Demandante, professor ou administrador, conforme escopo | `200`, array de DTOs |
| GET | `/demands/:id` | Proprietário; professor com acesso à prateleira/atribuição; administrador | `200`, DTO |
| PUT | `/demands/:id` | Demandante proprietário, estado `cadastrada` | `200`, DTO atualizado |
| DELETE | `/demands/:id` | Mesma regra de cancelamento, com motivo obrigatório | `200`, DTO cancelado |
| POST | `/demands/:id/analisar` | Professor; cadastrada e sem atribuição | `200`, DTO em análise |
| POST | `/demands/:id/devolver` | Professor atribuído; em análise e dentro do prazo | `200`, DTO cadastrado |
| POST | `/demands/:id/assumir` | Professor atribuído; em análise e dentro do prazo | `200`, DTO em produção |
| POST | `/demands/:id/concluir` | Professor atribuído; em produção | `200`, DTO concluído |
| POST | `/demands/:id/cancelar` | Demandante proprietário; cadastrada | `200`, DTO cancelado |

Parâmetro da listagem:

| `escopo` | Comportamento |
| --- | --- |
| Omitido ou vazio | Demandante: próprias; professor: prateleira; administrador: todas |
| `minhas` | Próprias do demandante ou atribuídas ao professor; administrador recebe `403` |
| `prateleira` | Professor: cadastradas e em produção; demais papéis recebem `403` |
| Outro valor | `400` |

Listas retornam `[]` quando vazias e são ordenadas por criação e ID decrescentes.
Não há filtros de período, pesquisa, paginação ou ordenação configurável nessas
rotas. Campos internos são rejeitados; área é opcional; cancelamento preserva o
registro. Corpos das ações, DTO completo e regras de expiração:
[demandas.md](demandas.md).

## Recomendações

| Método | Rota | Acesso | Resultado de sucesso |
| --- | --- | --- | --- |
| GET | `/teachers/:id/recommendations?limit=10` | Própria conta com perfil de professor; sem exceção administrativa para terceiros | `200`, objeto com `teacher_id` e `recommendations` |

`limit` aceita de 1 a 100, com padrão 10. Exemplo de resposta ilustrativa:

```json
{
  "teacher_id": 2,
  "recommendations": [
    {
      "demand_id": 42,
      "title": "Agendamento de salas",
      "score": 0.72,
      "matched_area": {"id":1,"name":"Desenvolvimento de Software","type":"técnica"}
    }
  ]
}
```

Ausência de correspondências produz `recommendations: []`. O score é uma
similaridade ponderada, não uma probabilidade. O endpoint retorna um resumo;
use `GET /demands/:id` para os detalhes. Critérios do ranking estão em
[arquitetura.md](arquitetura.md).

## Usuários administrativos

| Método | Rota | Acesso | Resultado de sucesso |
| --- | --- | --- | --- |
| GET | `/users/` | Administrador | `200`, lista dos modelos de usuário |
| POST | `/users/` | Administrador | `201`, `{"userID":3}` |

Exemplo de criação administrativa:

```json
{"name":"Nome","email":"cadastro@example.com","role":"demandante","foto_url":null}
```

O hook do modelo pode sobrescrever o papel conforme o e-mail/domínio. Essa rota
não cria perfil de demandante/professor nem vínculo Google. Um login posterior
com o mesmo e-mail pode retornar `409` por ausência do vínculo; o fluxo normal
do aplicativo é criar a conta pelo login e completar o perfil depois.

Esse CRUD ainda recebe e retorna o modelo GORM diretamente. A consulta inclui
`ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt`, `name`, `role`, `email` e `foto_url`;
não é o DTO de `/me`. Lista vazia pode ser `null`. Não há GET por ID, PUT ou
DELETE de usuários registrados. A validação e os erros ainda são os do handler
antigo, inclusive possíveis respostas `500` para conflitos do banco.

## Convenções e limites dos contratos

- `200`: consulta ou ação com resultado; `201`: criação; `204`: sucesso sem
  corpo. Um `204` não deve ser decodificado como JSON.
- Os endpoints principais usam `{"error":{"code":400,"message":"..."}}`.
  `code` é numérico e corresponde ao status HTTP.
- `400` indica entrada inválida; `401`, autenticação; `403`, acesso negado;
  `404`, recurso ausente; `409`, conflito; `500`, falha interna. Os módulos
  antigos de áreas/usuários não fazem todos esses mapeamentos.
- Login e `/me` enviam `Cache-Control: no-store`.
- IDs são numéricos. SIAPE é textual, incluindo zeros à esquerda.
- Não há um envelope JSON único para todas as listas nem uma convenção única
  de nomes. Use os DTOs e exceções documentados por módulo.
- Não há versionamento global por prefixo: o `/v2` pertence somente ao caminho
  de login. Nenhum endpoint Swagger/OpenAPI está registrado atualmente.

As rotas foram conferidas em [server/routes.go](../internal/server/routes.go)
e nos arquivos `*_routes.go` dos respectivos domínios.
