# Notificações e push Android

Implementação da API em 09/11/2026. A integração Flutter, o deploy e a
entrega real em um dispositivo ainda precisam ser realizados e validados.

## Responsabilidades

PostgreSQL armazena o histórico e a fila de entregas; a API determina os
destinatários. Firebase Cloud Messaging (FCM) apenas transporta o push. Não
é necessário usar Firestore, Firebase Authentication ou Cloud Functions.

- Alterações de estado notificam o demandante, inclusive cancelamento e
  devolução automática após expiração da análise. Edições sem mudança de
  estado não notificam. A gravação acontece na mesma transação da alteração.
- A rotina de recomendações roda ao iniciar e a cada 30 segundos, avaliando
  as dez primeiras demandas de cada professor ativo pelo mesmo ranking da
  consulta atual. Cada par professor/demanda gera um aviso uma única vez.
  O primeiro processamento também notifica recomendações já existentes.
  Embeddings da demanda e das áreas são necessários. Até 500 novos avisos
  são processados por execução; reinícios não duplicam os avisos.
- `POST /reports` registra uma denúncia e notifica todos os administradores
  existentes na mesma transação. Um professor pode denunciar cada demanda
  uma vez. Esta versão inclui cadastro e consulta, sem julgamento ou strikes.

## Contrato HTTP

Todas as rotas exigem `Authorization: Bearer <token da API>`.
O destinatário sempre vem da sessão, nunca do corpo enviado pelo aplicativo.

| Método | Rota | Resultado |
| --- | --- | --- |
| GET | `/notifications?limit=30&before=123` | Array de notificações próprias, ID decrescente |
| GET | `/notifications/unread-count` | `{"count":3}` |
| PATCH | `/notifications/:id/read` | `204`; preserva a primeira data de leitura |
| PATCH | `/notifications/read-all` | `204`; marca todas as próprias |
| PUT | `/notifications/devices` | `204`; registra/atualiza o dono do token FCM |
| DELETE | `/notifications/devices` | `204`; remove somente token pertencente à sessão |
| POST | `/reports` | Professor ativo; `201` com a denúncia |
| GET | `/reports?limit=30&before=123` | Administrador; array de denúncias |
| GET | `/reports/:id` | Administrador; detalhes da denúncia |

`before` é opcional e corresponde ao último ID da página anterior; `limit`
vai de 1 a 100. Lista vazia retorna `[]`. Notificação inexistente ou de outro
usuário retorna `404` ao tentar marcar como lida. Erros seguem o envelope
`{"error":{"code":400,"message":"..."}}`.

Exemplo de notificação:

```json
{
  "id": 12,
  "user_id": 4,
  "type": "demandaEmAnalise",
  "title": "Estado da demanda alterado",
  "body": "Site institucional: em análise.",
  "resource_type": "demand",
  "resource_id": 8,
  "created_at": "2026-11-09T15:00:00Z",
  "read_at": null
}
```

Tipos: `demandaEmAnalise`, `demandaEmProducao`, `demandaDevolvida`,
`demandaConcluida`, `demandaCancelada`, `demandaRecomendada`, `denunciaRecebida`.
`resource_type` é `demand` ou `report`. IDs são numéricos no HTTP.

Registro e remoção de dispositivo recebem `{"token":"TOKEN_FCM_DO_APARELHO"}`.
A remoção usa corpo JSON, sem token em URL. Registrar o mesmo token em outra
conta transfere seu vínculo; entregas antigas são descartadas se o dono mudar.

Denúncia recebe:

```json
{"demandaId":8,"motivo":"spam","descricao":"Descrição do problema encontrado."}
```

Motivos: `conteudoOfensivo`, `spam`, `foraDoEscopo`, `informacaoFalsa`,
`demandaDuplicada`, `outro`. Descrição obrigatória, até 5000 bytes. Só permite
denunciar demandas que o professor pode acessar. Repetição retorna `409`.

## Configurar a VPS

### Arquivo local e imagem Docker

O arquivo local fica em `run/secrets/firebase-service-account.json`.
Para executar Go diretamente na raiz do projeto,
use `GOOGLE_APPLICATION_CREDENTIALS=./run/secrets/firebase-service-account.json`
no `.env` e o `FCM_PROJECT_ID` correspondente à credencial.

O Dockerfile define o caminho `/run/secrets/firebase-service-account.json`, mas **não copia o JSON**. No painel da hospedagem, configure uma montagem de arquivo secreto no serviço
da API com destino `/run/secrets/firebase-service-account.json`. Se o painel
oferecer bind mount, envie o arquivo ao servidor por um canal privado e monte
esse arquivo no mesmo destino, somente leitura. Não
adicione `COPY run/` ao Dockerfile nem envie a credencial para o repositório.

### Ativação do envio

1. No projeto Firebase usado pelo aplicativo, habilitar a API Firebase Cloud
   Messaging HTTP v1. Criar uma credencial de conta de serviço com permissão
   para enviar mensagens nesse projeto.
2. Instalar o JSON da credencial como arquivo secreto do serviço/container,
   por exemplo `/run/secrets/firebase-service-account.json`. Não colocar no
   Git, no APK nem em uma pasta pública.
3. Configurar no painel da hospedagem:

   ```text
   FCM_PROJECT_ID=id-do-projeto-firebase
   GOOGLE_APPLICATION_CREDENTIALS=/run/secrets/firebase-service-account.json
   ```

   O caminho precisa existir dentro do container, não apenas na máquina host.
4. Publicar a API e reiniciar o serviço. A inicialização migra as tabelas
   `notifications`, `devices`, `deliveries` e `reports`.

Sem `FCM_PROJECT_ID`, a central HTTP funciona, mas o worker de push fica
desativado. Com projeto configurado e credencial ausente/inválida na carga,
a inicialização falha explicitamente. Permissões e projeto incorretos também
podem causar erros de envio depois da inicialização.

## Integração Flutter pendente

1. Adicionar `firebase_messaging` compatível com a versão de `firebase_core`
   do app e inicializar o Firebase também no modo API, apenas para Messaging.
2. Solicitar permissão de notificações no Android e obter o token FCM.
3. Após login na API, enviar o token a `PUT /notifications/devices` e atualizar
   o registro em `onTokenRefresh`. No logout, remover o dispositivo **antes**
   de descartar o JWT; invalidar o token FCM quando não for possível desvincular.
4. Implementar o repositório HTTP de notificações, habilitar o sino no modo
   API e consultar periodicamente enquanto houver uma sessão ativa. Consumir
   todas as páginas necessárias ou usar o endpoint de contagem para o badge.
5. Com o app em primeiro plano, tratar `onMessage` para atualizar a central
   e, se desejado, exibir um aviso. Em segundo plano, a mensagem inclui
   `notification`, permitindo ao Android exibir o push.
6. Tratar `getInitialMessage` e `onMessageOpenedApp`. Os campos `data` são
   strings: `notification_id`, `resource_type`, `resource_id`. Abrir o recurso
   somente depois de autenticar e consultar a API; ela valida o acesso.
7. Integrar o cadastro e as consultas de denúncias às rotas acima. A seleção
   dos administradores acontece no servidor, sem consulta de usuários pelo professor.

## Entregas e limites desta versão

O worker busca até 20 entregas a cada ciclo (intervalo de 5 segundos), com
timeout de 10 segundos por envio e 25 segundos por lote. Erros são tentados
até oito vezes, com espera crescente a partir de 60 segundos. Tokens que o
FCM informa como `UNREGISTERED` são removidos. Após oito tentativas a entrega
é encerrada; o histórico da notificação permanece disponível.

Somente dispositivos registrados no momento do evento recebem push. Login
posterior permite consultar o histórico, sem disparar notificações antigas.
Há proteção contra workers concorrentes, mas uma queda entre a aceitação
pelo FCM e o commit pode repetir um envio. A tag Android estável reduz avisos
duplicados na bandeja. Não há garantia de entrega imediata nem recibo de leitura
do push. O botão “forçar parada” do Android impede o recebimento até reabrir o app.


Referências oficiais:
[FCM no Flutter](https://firebase.google.com/docs/cloud-messaging/flutter/get-started),
[recebimento](https://firebase.google.com/docs/cloud-messaging/flutter/receive-messages),
[envio HTTP v1](https://firebase.google.com/docs/cloud-messaging/send/v1-api).
