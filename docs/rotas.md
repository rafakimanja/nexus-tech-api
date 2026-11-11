# Rotas da API

[Introdução](../README.md) · [Arquitetura](arquitetura.md)

Abra **`/docs`** na API em execução para explorar todas as rotas, informar o
JWT, preencher parâmetros e executar chamadas. A especificação legível por
ferramentas está em **`/openapi.json`**. Os caminhos abaixo são relativos à
origem da API; coleções de usuários, áreas, perfis e demandas usam `/` final.

Envie JSON com `Content-Type: application/json`. Exceto nas rotas públicas,
envie `Authorization: Bearer <access_token>`, usando o JWT retornado pelo login
da API, não o ID token Google. A interface faz chamadas reais e pode modificar
dados; execute operações de escrita em ambiente apropriado.

## Sessão e usuários

| Método e rota | Acesso e finalidade |
| --- | --- |
| `GET /`, `GET /health` | Público; identificação e diagnóstico |
| `GET /login`, `GET /auth/google`, `GET /auth/google/callback` | Fluxo OAuth manual no navegador |
| `POST /login/v2` | Público; recebe `{"idToken":"..."}` do Google e retorna `access_token` |
| `GET /me` | Conta autenticada, inclusive bloqueada; perfil e motivo do bloqueio |
| `GET /users/`, `POST /users/` | Administrador; listar ou criar usuários |

## Perfis e áreas

| Método e rota | Acesso e finalidade |
| --- | --- |
| `GET /applicants/`, `GET /teachers/` | Administrador; listar perfis |
| `GET /applicants/:id`, `GET /teachers/:id` | Próprio usuário ou administrador |
| `POST /applicants/`, `POST /teachers/` | Demandante ou professor cria seu próprio perfil |
| `PUT /applicants/:id`, `PUT /teachers/:id` | Proprietário atualiza seu perfil completo |
| `GET /areas/`, `GET /areas/:id` | Conta autenticada; consultar áreas |
| `POST /areas/`, `PUT /areas/:id`, `DELETE /areas/:id` | Conta autenticada; manter áreas |
| `GET /areas-teachers/professor/:id` | Próprio usuário ou administrador; consultar vínculos |
| `PUT /areas-teachers/professor/:id`, `PUT /teachers/:id/areas` | Professor proprietário; substituir vínculos |
| `GET /teachers/:id/recommendations?limit=10` | Professor proprietário; recomendações, limite 1 a 100 |

O perfil de professor requer SIAPE textual de sete dígitos e áreas técnicas e de
interesse. Os tipos de demandante aceitos são `empresa`, `orgaoPublico`, `ong`,
`instituicaoEducacional`, `pessoaFisica` e `academico`. Veja exemplos de JSON
na interface interativa.

## Demandas

| Método e rota | Acesso e finalidade |
| --- | --- |
| `GET /demands/?escopo=minhas` | Lista conforme papel e escopo: `minhas` ou `prateleira` |
| `POST /demands/` | Demandante cria uma demanda |
| `GET /demands/:id` | Proprietário, professor com acesso ou administrador |
| `PUT /demands/:id` | Proprietário edita demanda cadastrada |
| `DELETE /demands/:id` ou `POST /demands/:id/cancelar` | Proprietário cancela; JSON com `motivoCancelamento` |
| `POST /demands/:id/analisar` | Professor reserva demanda cadastrada |
| `POST /demands/:id/devolver` | Professor atribuído devolve análise |
| `POST /demands/:id/assumir` | Professor atribuído inicia produção |
| `POST /demands/:id/concluir` | Professor atribuído conclui; JSON com `descricaoSolucao` |

Criação e edição recebem `titulo`, `descricao`, `publicoAlvo`, `impacto` e
`areaId` opcional. Cancelar preserva o registro. Ações fora do estado permitido
retornam `409`.

## Notificações e denúncias

| Método e rota | Acesso e finalidade |
| --- | --- |
| `GET /notifications?limit=30&before=123` | Notificações da conta, em ordem decrescente |
| `GET /notifications/unread-count` | Quantidade não lida |
| `PATCH /notifications/:id/read`, `PATCH /notifications/read-all` | Marcar como lida(s) |
| `PUT /notifications/devices`, `DELETE /notifications/devices` | Registrar ou remover `{"token":"TOKEN_FCM"}` |
| `POST /reports` | Professor ativo denuncia demanda acessível |
| `GET /reports?status=pendente&limit=30&before=123` | Administrador lista denúncias |
| `GET /reports/:id` | Administrador ou professor autor consulta denúncia |
| `PATCH /reports/:id/judgment` | Administrador registra decisão final e justificativa |

Denúncia recebe `{"demandaId":1,"motivo":"spam","descricao":"..."}`.
Motivos: `conteudoOfensivo`, `spam`, `foraDoEscopo`, `informacaoFalsa`,
`demandaDuplicada`, `outro`. Cada professor pode denunciar uma vez cada demanda.
Julgamento recebe `{"decisao":"procedente","justificativa":"..."}`;
decisão também pode ser `improcedente`. Procedência cancela a demanda e aplica
até um strike por demanda. Denúncia julgada novamente retorna `409`.

## Respostas

Sucesso normalmente usa `200`, `201` ou `204` (sem corpo). Erros das rotas
principais seguem `{"error":{"code":400,"message":"..."}}`; o callback
OAuth manual tem formato próprio. `400` indica entrada inválida, `401` falha de
autenticação, `403` acesso negado, `404` recurso ausente, `409` conflito e
`500` falha interna. Alguns módulos legados ainda não padronizam todos os erros.
