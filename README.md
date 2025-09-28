# Summary Bot

Telegram бот на Go для создания кратких выжимок из постов в каналах с использованием LLM (GPT).

## Функциональность

- Подписка на Telegram каналы
- Автоматическое создание выжимок из новых постов
- Настройка формата и языка выжимок
- Поддержка многих пользователей

## Требования

- Go 1.16+
- PostgreSQL
- Redis (опционально, для кэширования)
- Telegram Bot Token (получить у @BotFather)
- OpenAI API ключ

## Установка и запуск

### 1. Клонирование репозитория

```bash
git clone https://github.com/nzyazin/summary-bot.git
cd summary-bot
```

### 2. Настройка конфигурации

Отредактируйте файл `config/config.yaml`, указав ваши данные:

```yaml
telegram:
  token: "YOUR_TELEGRAM_BOT_TOKEN" # Замените на ваш токен от BotFather
  
llm:
  api_key: "YOUR_OPENAI_API_KEY" # Замените на ваш API ключ OpenAI
```

### 3. Создание базы данных

```bash
psql -U postgres -c "CREATE DATABASE summary_bot;"
psql -U postgres -d summary_bot -f migrations/init.sql
```

### 4. Запуск бота

```bash
go run cmd/main.go
```

## Использование

1. Найдите бота в Telegram и отправьте команду `/start`
2. Используйте команду `/subscribe` и отправьте ссылку на канал для подписки
3. Бот будет автоматически создавать выжимки из новых постов
4. Используйте `/settings` для настройки формата и языка выжимок
