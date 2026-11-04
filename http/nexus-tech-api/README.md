# Requisições manuais

Configure `url` com o endereço da API. Após executar o login, defina a variável
local `access_token` com o JWT retornado em `access_token`. Não salve tokens reais
nos exemplos versionados.

Para substituir as áreas do professor, defina `professor_id` com o ID numérico
do professor autenticado. O corpo substitui a lista inteira; os IDs das áreas
devem existir. O cadastro de professor também exige uma área técnica e uma de
interesse, com IDs distintos.

Os contratos atuais estão em [autenticacao.md](../../docs/autenticacao.md),
[perfis.md](../../docs/perfis.md) e [demandas.md](../../docs/demandas.md).
