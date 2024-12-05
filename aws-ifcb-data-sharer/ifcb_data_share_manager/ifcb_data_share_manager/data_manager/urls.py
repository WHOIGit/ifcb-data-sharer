from django.urls import path
from .views import DatasetListView, DatasetDetailView

app_name = "data_manager"

urlpatterns = [
    path("", DatasetListView.as_view(), name="datasets_list"),
    path("<int:pk>/", DatasetDetailView.as_view(), name="dataset_detail"),
]
