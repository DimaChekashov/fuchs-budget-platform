from django.db import models

class Transaction(models.Model):
    """Модель транзакции"""
    id = models.AutoField(primary_key=True)
    value = models.DecimalField(
        max_digits=10, 
        decimal_places=2,
        help_text="Сумма транзакции"
    )
    title = models.CharField(max_length=200, help_text="Название транзакции")
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    class Meta:
        ordering = ['-created_at']
        verbose_name = "Транзакция"
        verbose_name_plural = "Транзакции"

    def __str__(self):
        return f"{self.title}: {self.value}"