from django.contrib import admin
from .models import Transaction

@admin.register(Transaction)
class TransactionAdmin(admin.ModelAdmin):
    list_display = ['id', 'title', 'value', 'created_at']
    list_filter = ['created_at']
    search_fields = ['title']