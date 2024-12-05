import contextlib

from django.apps import AppConfig
from django.utils.translation import gettext_lazy as _


class UsersConfig(AppConfig):
    name = "ifcb_data_share_manager.users"
    verbose_name = _("Users")

    def ready(self):
        with contextlib.suppress(ImportError):
            import ifcb_data_share_manager.users.signals  # noqa: F401
