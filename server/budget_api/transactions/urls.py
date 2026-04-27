from django.urls import path
from . import views

urlpatterns = [
    path('transactions/', views.get_all_transactions, name='transactions-list'),
    path('transactions/create/', views.create_transaction, name='transaction-create'),
    path('transactions/<int:pk>/', views.get_transaction, name='transaction-detail'),
    path('transactions/<int:pk>/delete/', views.delete_transaction, name='transaction-delete'),
]