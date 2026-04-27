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