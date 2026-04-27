from .models import Transaction

class TransactionService:
    """Сервис для работы с транзакциями"""
    
    @staticmethod
    def create_transaction(title: str, value: float) -> Transaction:
        """Создать новую транзакцию с бизнес-логикой"""
        
        if value <= 0:
            raise ValueError("Value must be positive")
        
        if len(title) < 3:
            raise ValueError("Title too short")
        
        return Transaction.objects.create(title=title, value=value)
    
    @staticmethod
    def get_total_amount() -> float:
        """Подсчитать общую сумму всех транзакций"""
        total = Transaction.objects.aggregate(
            total=models.Sum('value')
        )['total'] or 0
        return float(total)