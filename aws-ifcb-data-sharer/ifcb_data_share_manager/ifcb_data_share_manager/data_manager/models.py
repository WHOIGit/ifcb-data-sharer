from django.db import models
from django.utils import timezone
from ..users.models import User


class Dataset(models.Model):
    # Datasets to sync with ifcbdb

    name = models.CharField(max_length=100, unique=True, db_index=True)
    title = models.CharField(max_length=100, db_index=True)
    user = models.ForeignKey(User, related_name="datasets", on_delete=models.CASCADE)
    created_date = models.DateTimeField(default=timezone.now)

    class Meta:
        ordering = ["name"]

    def __str__(self):
        return self.name


class Bin(models.Model):
    # IFCB Bins attached to each dataset
    # data is retrieved from Dynamodb master list
    pid = models.CharField(max_length=100, db_index=True)
    s3_key_root = models.CharField(max_length=100, blank=True, db_index=True)
    dataset = models.ForeignKey(Dataset, related_name="bins", on_delete=models.CASCADE)

    class Meta:
        ordering = ["pid"]

    def __str__(self):
        return self.pid
