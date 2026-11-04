# Documentação da Nexus Tech API

Referência do código deste repositório, revisada em 04/11/2026.

A API conecta demandantes que publicam necessidades a professores que analisam,
assumem e concluem demandas. Ela mantém identidades Google, sessões próprias,
perfis, áreas e o ciclo de vida das demandas em PostgreSQL. Também oferece
recomendações por similaridade de texto usando Ollama e pgvector.

## Por onde começar

| Objetivo | Documento |
| --- | --- |
| Preparar o ambiente, executar e testar | [Configuração e execução](configuracao.md) |
| Entender componentes, dados e fluxos | [Arquitetura e fluxos](arquitetura.md) |
| Consultar métodos, URLs e permissões | [Referência de rotas](rotas.md) |
| Integrar Google, JWT e `/me` | [Autenticação](autenticacao.md) |
| Cadastrar e editar demandantes e professores | [Perfis](perfis.md) |
| Publicar, consultar e movimentar demandas | [Demandas](demandas.md) |
| Consultar exemplos de requisições manuais | [Coleção HTTP](../http/nexus-tech-api/README.md) |
| Acompanhar trabalho restante | [Pendências de integração](pendencias-integracao.md) |

[auditoria.md](auditoria.md) é um registro histórico de 02/11/2026. Seus achados
não representam necessariamente o estado atual: várias correções foram
implementadas depois da auditoria.

## Sequência básica de integração

1. Obter um ID token Google e enviá-lo em `POST /login/v2`.
2. Guardar o `access_token` retornado e enviá-lo como `Authorization: Bearer JWT`.
3. Consultar `GET /me` para identificar papel, bloqueio e situação do perfil.
4. Completar o perfil com `POST /applicants/` ou `POST /teachers/`.
5. Publicar demandas como demandante ou consultar a prateleira como professor.
6. Usar as ações específicas de demanda para analisar, devolver, assumir,
   concluir ou cancelar; não enviar mudanças de estado no formulário de edição.

Os IDs são numéricos. Há convenções JSON diferentes entre módulos: demandas usam
nomes como `publicoAlvo`, enquanto perfis usam nomes como `photo_url`. Use o
contrato de cada recurso, sem aplicar uma conversão global de nomes no cliente.

## Escopo atual

Autenticação, perfis, áreas, demandas, associações docentes e recomendações têm
rotas implementadas. A API ainda não oferece chat, notificações, upload de
arquivos, denúncias, republicação nem gestão administrativa completa. Não há
refresh token, WebSocket, SSE ou contrato geral de paginação.

Esta documentação descreve o comportamento atual, inclusive limitações. A
existência de um endpoint não significa que ele faça parte do fluxo mínimo do
aplicativo ou que seus testes de integração tenham sido executados no ambiente.
