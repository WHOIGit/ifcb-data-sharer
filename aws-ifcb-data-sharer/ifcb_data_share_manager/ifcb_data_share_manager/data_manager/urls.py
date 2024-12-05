from django.urls import path
from .views import DatasetListView

app_name = "data_manager"

urlpatterns = [
    path("", DatasetListView.as_view(), name="datasets_list"),
]
