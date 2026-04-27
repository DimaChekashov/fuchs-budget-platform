# Backend

Run server:
```bash
source ../venv/bin/activate

python manage.py runserver
```

Commands
```bash
# Показать все URL маршруты
python manage.py show_urls

# Создать миграции после изменений в моделях
python manage.py makemigrations

# Выполнить миграции
python manage.py migrate

# Запустить интерактивную оболочку Django
python manage.py shell
```

### Domain-based architecture
```bash
my_project/
├── manage.py
├── core/                # Настройки проекта, WSGI, ASGI
│   ├── settings/
│   └── urls.py          # Главный routing (включает urls приложений)
├── apps/                # Папка для всех внутренних приложений
│   ├── users/           # Сущность 1: Пользователи
│   │   ├── models.py
│   │   ├── serializers.py
│   │   ├── views.py
│   │   ├── services.py   # Бизнес-логика (чтобы не раздувать views)
│   │   └── urls.py
│   ├── products/        # Сущность 2: Товары
│   │   ├── models.py
│   │   ├── serializers/  # Если сущность сложная, делаем папку
│   │   │   ├── base.py
│   │   │   └── nested.py
│   │   └── views.py
│   ├── orders/          # Сущность 3: Заказы
│   ├── payments/        # Сущность 4: Платежи
│   └── catalog/         # Сущность 5: Категории/Теги
├── common/              # Общие утилиты, миксины, базовые классы
│   ├── models.py        # Напр. BaseModel с createdAt/updatedAt
│   └── permissions.py
├── requirements.txt
└── .env 
```