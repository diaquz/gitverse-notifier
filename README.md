
config example
```
repositories:
  default:
    event:
      default:
        notify_jira: true
        update_pull_request: true
    
  bks-ruby/control-objects:
    event:
      pull_request_created:
        notify_telegram: true
        
```

settings
```
settings:
  telegram:
    token: ...
    users:
      gitverse_email: telegram_tag
  jira:
    url: ...
    token: ...
```


telegram message template?
telegram support some kind of formating i think
```
pull_request.template

[${repository}] New pull request [${pull_request.title}](${pull_request.url}) by ${tg_mapping(pull_request.login)}
```
