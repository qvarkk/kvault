# Участие в разработке — Фронтенд (kvault-frontend)

Спасибо за интерес к kvault. Этот документ описывает работу над **фронтендом** (веб-интерфейс). Репозиторий бэкенда: [gitverse.ru/qvarkk/kvault](https://gitverse.ru/qvarkk/kvault).

## Стек

Vue 3 (Composition API) · TypeScript · Vite · Tailwind CSS v4 · Pinia · Vue Router · shadcn-style компоненты на reka-ui

## Локальное окружение

```bash
# Установить зависимости
npm ci

# Скопировать и заполнить конфиг
cp .env.example .env

# Запустить dev-сервер (Vite)
npm run dev
```

Требуемые переменные `.env`:

```
VITE_API_URL=http://localhost:6767/api/v1
VITE_APP_NAME=kvault
```

Для работы интерфейса нужен запущенный бэкенд — см. репозиторий [gitverse.ru/qvarkk/kvault](https://gitverse.ru/qvarkk/kvault).

## Полезные команды

```bash
npm run dev          # dev-сервер
npm run build        # проверка типов + сборка
npm run type-check   # только vue-tsc
npm run build-only   # сборка без проверки типов
npm run preview      # предпросмотр прод-сборки
```

## Как внести вклад

1. Создайте ветку от `main`.
2. Внесите изменения; держите коммиты сфокусированными, пишите осмысленные сообщения.
3. Перед PR прогоните проверку типов и сборку: `npm run build`.
4. Откройте pull request с описанием сути и причины изменений.

## Правила по коду

- Импорты из `src/` через алиас `@/`.
- Все видимые пользователю строки — через i18n (`useI18n().t(...)`), без хардкода текста. Добавляйте переводы во все локали (`en`, `ru`, `ja`).
- Сервисы (`src/services/`) — тонкие обёртки над axios, по одному файлу на ресурс API.
- Списки используют существующие composables (`useEntities`, `useInfiniteEntities`) — не дублируйте логику загрузки.
- UI собирайте из примитивов `src/components/ui/`; формы — на vee-validate + zod.

## Сообщения об ошибках

Для багов и предложений создавайте issue с описанием, шагами воспроизведения и ожидаемым поведением. Уязвимости — не в публичные issue, см. [SECURITY.md](./SECURITY.md).

Участвуя в проекте, вы соглашаетесь соблюдать [Кодекс поведения](./CODE_OF_CONDUCT.md).
