import boto3
import environ
from django.shortcuts import render
from django.views.generic import ListView, DetailView
from django.contrib.auth.mixins import LoginRequiredMixin

from .models import Dataset

env = environ.Env()

AWS_ACCESS_KEY_ID = env("DJANGO_AWS_ACCESS_KEY_ID")
AWS_SECRET_ACCESS_KEY = env("DJANGO_AWS_SECRET_ACCESS_KEY")
AWS_DEFAULT_REGION = env("AWS_DEFAULT_REGION")
DYNAMO_TABLE = "ifcb-data-sharer-bins"


def query_dynamodb(user):
    dynamodb_client = boto3.client(
        "dynamodb",
        aws_access_key_id=AWS_ACCESS_KEY_ID,
        aws_secret_access_key=AWS_SECRET_ACCESS_KEY,
        region_name=AWS_DEFAULT_REGION,
    )

    response = dynamodb_client.query(
        TableName=DYNAMO_TABLE,
        KeyConditionExpression="username = :username",
        ExpressionAttributeValues={":username": {"S": user.username}},
    )
    return response["Items"]


class DatasetListView(LoginRequiredMixin, ListView):
    model = Dataset
    template_name = "data_manager/dataset_list.html"
    context_object_name = "datasets"

    def get_queryset(self):
        queryset = super().get_queryset()
        return queryset.filter(user=self.request.user)


class DatasetDetailView(LoginRequiredMixin, DetailView):
    model = Dataset
    context_object_name = "dataset"
    template_name = "data_manager/dataset_detail.html"

    def get_context_data(self, **kwargs):
        obj = self.get_object()
        context = super(DatasetDetailView, self).get_context_data(**kwargs)

        # get all available Bins for this user from Dynamodb and format for display
        response = query_dynamodb(obj.user)
        print(response)
        all_bins = [item["pid"]["S"] for item in response]

        print(all_bins)
        context.update({"all_bins": all_bins})
        return context
