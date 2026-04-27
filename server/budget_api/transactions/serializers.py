from rest_framework import serializers
from .models import Transaction

class TransactionSerializer(serializers.ModelSerializer):
    """Сериализатор для транзакций"""
    
    class Meta:
        model = Transaction
        fields = ['id', 'value', 'title', 'created_at', 'updated_at']
        read_only_fields = ['id', 'created_at', 'updated_at']

class CreateTransactionSerializer(serializers.Serializer):
    """Сериализатор для создания транзакции"""
    title = serializers.CharField(max_length=200)
    value = serializers.DecimalField(max_digits=10, decimal_places=2)