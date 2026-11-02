# Autenticação do MVP

## Configuração e sessão

- `GOOGLE_CLIENT_ID`: audiência do ID token Google enviado pelo aplicativo.
- `JWT_SECRET_KEY`: segredo aleatório com pelo menos 32 caracteres. A API rejeita a inicialização se inválido. O nome antigo `JWT_SECRECT_KEY` continua aceito somente quando o novo não está definido. Manter o mesmo valor durante a troca de nome preserva tokens existentes; trocar o valor invalida todos os tokens anteriores.
- Tokens novos duram **12 horas** (43.200 segundos), assinados com HS256. Tokens antigos mantêm sua expiração original.
- Não há refresh token nem revogação individual neste MVP. Após expirar, o app obtém outro ID token Google e faz login novamente. Logout local remove o token do armazenamento seguro do app; uma cópia do token ainda pode funcionar até expirar.

## POST /login/v2

Entrada: `{"idToken":"ID_TOKEN_GOOGLE"}`. Enviar o ID token Google, não um token Firebase.

Resposta 200:

```json
{
  "access_token": "JWT",
  "token_type": "Bearer",
  "expires_in": 43200,
  "expires_at": "2026-11-03T00:00:00Z",
  "user_id": 42,
  "user": {
    "id": 42,
    "name": "Nome",
    "email": "usuario@example.com",
    "role": "demandante",
    "photo_url": null,
    "profile_complete": false,
    "blocked": false
  }
}
```

`expires_at` é UTC e corresponde ao `exp` do JWT. O cabeçalho `Authorization` e o campo `user_id` permanecem para compatibilidade. Respostas de login e `/me` usam `Cache-Control: no-store`.

O vínculo usa `(google, subject)`. A criação de conta e identidade é transacional, com serialização para logins concorrentes e índice único no vínculo. O e-mail precisa estar verificado pelo Google. Contas preexistentes com o mesmo e-mail e sem o vínculo correspondente retornam **409**: precisam de vinculação controlada/migração, não há associação automática por e-mail. Contas vinculadas continuam identificadas pelo subject, mesmo que o e-mail Google mude.

O papel é definido pelo servidor usando as regras existentes de criação de usuários; novos e-mails fora das regras especiais recebem `demandante`. A política institucional e a configuração de administradores ainda precisam da revisão de produto prevista nas próximas etapas.

## GET /me

Enviar `Authorization: Bearer JWT`. Retorna `{"user": {...}}` com o mesmo DTO do login e estado atualizado no banco.

- Admin: cadastro considerado completo.
- Professor: completo se há nome, perfil com SIAPE de sete dígitos e áreas dos dois tipos; um perfil existente inativo é bloqueado (`professor_inactive`).
- Demandante: completo se há nome e perfil com telefone, tipo válido, CPF/CNPJ e endereço; três ou mais strikes bloqueiam a sessão (`strike_limit`).
- Perfil ainda inexistente: cadastro incompleto, permitindo seu preenchimento.
- Papel desconhecido: bloqueado (`invalid_role`).

Esses critérios refletem os campos atuais da API e deverão acompanhar a futura revisão do cadastro. `block_reason` é retornado somente quando bloqueado.

As rotas protegidas verificam o estado atual da conta a cada requisição, sem confiar no papel antigo do JWT. Usuários bloqueados recebem 403, mas podem consultar `/me` com token ainda válido para o app explicar o bloqueio. Contas removidas recebem 401. Isso não substitui a autorização por recurso dos CRUDs, que permanece uma etapa separada.

## Erros

Formato: `{"error":{"code":401,"message":"..."}}`.

- 400: entrada inválida ou `idToken` ausente.
- 401: identidade Google inválida, token expirado/inválido ou conta removida.
- 403: conta bloqueada.
- 409: conta exige vinculação de identidade.
- 500: falha interna; o app pode permitir nova tentativa.

Na inicialização, a migração cria o índice único de identidade. Duplicidades antigas impedem a inicialização e precisam ser analisadas antes da migração; nenhuma conta é mesclada ou removida automaticamente.
