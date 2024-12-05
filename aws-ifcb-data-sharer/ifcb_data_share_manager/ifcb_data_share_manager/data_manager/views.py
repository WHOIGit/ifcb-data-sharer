from django.shortcuts import render
from django.views.generic import ListView
from django.contrib.auth.mixins import LoginRequiredMixin

from .models import Dataset


class DatasetListView(LoginRequiredMixin, ListView):
    model = Dataset
    template_name = "data_manager/dataset_list.html"
    context_object_name = "datasets"

    def get_queryset(self):
        queryset = super().get_queryset()
        return queryset.filter(user=self.request.user)
