from django.contrib import admin
from .models import *

# Register your models here.


class DatasetAdmin(admin.ModelAdmin):
    list_display = (
        "name",
        "title",
        "user",
    )
    list_filter = ("user",)


class BinAdmin(admin.ModelAdmin):
    search_fields = ["pid"]
    list_display = ("pid", "dataset", "s3_key_root")
    list_filter = ("dataset",)


admin.site.register(Dataset, DatasetAdmin)
admin.site.register(Bin, BinAdmin)
