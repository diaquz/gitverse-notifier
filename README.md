# gitverse-notifier

Сервис для обработки вебхуков gitverse, создан для личных нужд и от безделья, частично навайбкожен с cursor.

## Запуск

### Локально

```bash
make init_configs
# нужно указать переменные в config.yml JIRA_TOKEN и тд
make clean & make build
./build/gitverse_notifier -f config.yml
```

### Docker

```bash
make init_configs
# нужно указать переменные в config.yml JIRA_TOKEN и тд
make docker
docker compose up -d
```

## Конфигурация

Конифиги можно укзать в config.yml или через переменные окружения.
Базовые конфиги:

| Ключ | Описание | Значение по умолчанию |
|---|---|---|
| `CONFIGS_DIR_PATH` | Путь к каталогу с конфигами | `configs/` |
| `TEMPLATES_DIR_PATH` | Путь к каталогу с шаблонами | `templates/` |
| `TEMPLATES_PATTERN` | Патерн, по которому происходит поиск шаблонов | `**/*.tmp` |
| `GITVERSE_BASE_URL` | Ссылка на гитверс для использования в уведомлениях | — |
| `BIND_HOST` | Хост | `0.0.0.0` |
| `HTTPD_PORT` | Порт | `9001` |
| `BASIC_AUTH_USER` | Пользователь для базовой аутентификации | — |
| `BASIC_AUTH_PASSWORD` | Пароль для базовой аутентификации | — |
| `LOG_LEVEL` | Уровень логирования | `INFO` |
| `LOG_FILE_NAME` | Имя файла с логами | `gitverse-notifier.log` |
| `LOG_MAX_SIZE` | Максимальный размер файла с логами (МБ) | `15` |
| `LOG_MAX_AGE` | Максимальный возвраст логов (дней)  | `7` |
| `LOG_FORMAT_JSON` | Форматировать ли логи в JSON  | `false` |
| `LOG_REQUESTS` | Логировать ли body для всех входящих событий  | `false` |
| `JIRA_URL` | URL для Jira | — |
| `JIRA_TOKEN` | Токен для Jira | — |
| `JIRA_USERNAME` / `JIRA_PASSWORD` | Логин и пароль для джиры - альтернатива токену | — |
| `TELEGRAM_BOT_TOKEN` | Токен Telegram-бота (опционально) | — |
| `TELEGRAM_CHAT_ID` | ID чата или канала для уведомлений (опционально) | — |
| `TELEGRAM_THREAD_ID` | ID треда в чате | — |
| `TELEGRAM_PROXY_URL` | HTTP прокси для Telegram (опционально) | — |
| `TELEGRAM_PARSE_MODE` | Режим парсинга сообщения `Markdown` / `MarkdownV2` / `HTML` | `MarkdownV2` |
| `GITVERSE_BASE_URL` | Ссылка на гитверс для использования в уведомлениях | — |
| `GITVERSE_API_URL` | Адрес публичного API gitverse | — |
| `GITVERSE_API_TOKEN` | Токен для публичного API gitverse | — |
| `CACHE_TTL` | Время кеширования информации о PR (в минутах) | `60` |
| `EVENT_QUEUE_SIZE` | Размер очереди событий | `256` |
| `EVENT_WORKER_COUNT` | Кол-во воркеров (горутин), разбирающих очередь событий | `8` |
| `ENABLE_PPROF` | Доступен ли API для pprof | `false` |

## Конфигурация репозиториев

Файлы с настройками репозиториев лежат в `configs/repositories/*.yml`, для примера есть `configs/repositories/default.yml`.

```yaml
repository: any
# Коды проектов Jira
allowed_jira_projects:
  - JIRA
  - TEST

# События, которые нужно группировать перед обработкой
event_groups:
  # Пример:
  #   Так как gitverse не присылает имя ревьювера в review_requested, а так же не имеет отдельного ивента для удаления ревьювера,
  #   то мы группируем все запросы на ревью в течение 60 секунд и отправляем уведомление только для последнего
  - key: events_group # Ключ группы
    group_by: group.pull_request # Способ группировки событий (имя репозитория + номер PR)
    strategy: strategy.use_last_event # Стратегия обработки событий - использовать последний ивент
    event: pull_request.review_requested # Собираемый тип событий
    size: 99 # ограничение по кол-ву событий в группе
    ttl: 60 # Время жизни группы, в секундах

action_rules:
  - on: pull_request.opened # Код события
    action: jira.comment_issue # Код действия
    template: jira/pr_opened # Название шаблона

  - on: pull_request.opened
    action: jira.mention_at_web # Web-ссылка на PR в задаче Jira

  - on: pull_request.opened
    action: telegram.notify
    template: telegram/pr_opened

  - on: branch.push
    branch: main
    action: telegram.notify
    template: telegram/branch_push

  - on: pull_request.comment
    action: telegram.notify
    skip_empty: true # gitverse бывает присылает события с пустым сообщением, опция для их пропуска
    template: telegram/pr_comment
```

### Список событий

Внутри приложения есть ряд типов событий, которые могут использоваться в action->on.
События из gitverse мапятся в них на основании значений заголовком X-Gitverse-Event и X-Gitverse-Event-Type (и иногда поля action).

| Событие | Описание |
|---|---|
| `pull_request.opened` / `.edited` / `.closed` / `synchronized` | Жизненый цикл PR |
| `pull_request.review_requested` | Запрос ревью PR |
| `pull_request.review_approved` | PR одобрен |
| `pull_request.review_rejected` | PR отклонен |
| `pull_request.review_comment` | Комментарий для PR (но сам комментарий не передается) |
| `pull_request.comment` | Комментарий для PR (с текстом) |
| `branch.push` | В ветку запушены изменения |
| `branch.created` / `branch.deleted` | Ветка создана / удалена |
| `cicd.status` | Изменение состояния CICD |


### Список действий

| Действие | Описание |
|---|---|
| `jira.comment_issue` | Создает комментарий для задачи в Jira |
| `jira.mention_at_web` | Добавляет web-ссылку (mentioned in) на PR к задаче в Jira |
| `telegram.notify` | Отправляет уведомление в Telegram |
| `utils.log` | Пишет событие в лог (для отладки) |


## Шаблоны

Файлы внутри `templates/**/*.tmpl` используются как go template.
Регистрируются с именами вида:
- `templates/jira/pr_opened.tmpl` -> `jira/pr_opened`
- `templates/telegram/pr_comment.tmpl` -> `telegram/pr_comment`

Имена шаблонов можно указываться в action->template.

Доступные хелперы: `tgMention`, `tgMentions`, `tgLink`, `tgIssueURLs`,  `jiraLink`, `jiraIssueURLs`, `IssueURL`, `tgIssueURLs`, `jiraIssueURLs`, `join`, `trim`, `jiraEscape`.
Пример шаблона

```
{{jiraLink .Sender.Name .Sender.URL}} открыл запрос на слияние{{if .PullRequest.Number}} #{{.PullRequest.Number}}{{end}}: {{jiraLink .PullRequest.Title .PullRequest.URL}} ({{jiraLink .Repository .RepositoryURL}})
{{- if .Branch}}
Ветка: {{jiraLink .Branch .BranchURL}}
{{- end}}
```

### Как парсятся номера задач Jira

- Из заголовка PR — любые вхождения вида `ABC-123`
- Из тела PR — любые вхождения вида `ABC-123`
- Из имени ветки (branch или ref)
- Задачи фильтруются по проектам, указанным в `allowed_jira_projects` из настроек репозитория

## В планах

- Поправить удаление группированных событий, если очередь полная 
- Написать тесты
