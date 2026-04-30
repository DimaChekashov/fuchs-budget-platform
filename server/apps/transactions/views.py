from rest_framework.decorators import api_view
from rest_framework.response import Response
from rest_framework import status
from .models import Transaction
from .serializers import TransactionSerializer, CreateTransactionSerializer
from .services import TransactionService

@api_view(['GET'])
def get_all_transactions(request):
    """Получить все транзакции"""
    transactions = Transaction.objects.all()
    serializer = TransactionSerializer(transactions, many=True)
    return Response(serializer.data)

@api_view(['POST'])
def create_transaction(request):
    """Создать транзакцию"""
    serializer = CreateTransactionSerializer(data=request.data)
    
    if serializer.is_valid():
        # Используем сервис для бизнес-логики
        transaction = TransactionService.create_transaction(
            title=serializer.validated_data['title'],
            value=serializer.validated_data['value']
        )
        
        # Возвращаем созданную транзакцию
        response_serializer = TransactionSerializer(transaction)
        return Response(response_serializer.data, status=status.HTTP_201_CREATED)
    
    return Response(serializer.errors, status=status.HTTP_400_BAD_REQUEST)

@api_view(['GET'])
def get_transaction(request, pk):
    """Получить одну транзакцию"""
    try:
        transaction = Transaction.objects.get(pk=pk)
    except Transaction.DoesNotExist:
        return Response(
            {'error': 'Transaction not found'}, 
            status=status.HTTP_404_NOT_FOUND
        )
    
    serializer = TransactionSerializer(transaction)
    return Response(serializer.data)

@api_view(['DELETE'])
def delete_transaction(request, pk):
    """Удалить транзакцию"""
    try:
        transaction = Transaction.objects.get(pk=pk)
        transaction.delete()
        return Response(
            {'message': 'Transaction deleted'}, 
            status=status.HTTP_204_NO_CONTENT
        )
    except Transaction.DoesNotExist:
        return Response(
            {'error': 'Transaction not found'}, 
            status=status.HTTP_404_NOT_FOUND
        )