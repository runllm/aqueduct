from aqueduct.integrations.integration import Integration, IntegrationInfo


class K8sIntegration(Integration):
    """
    Class for K8s integration.
    """

    def __init__(self, metadata: IntegrationInfo):
        self._metadata = metadata

    def describe(self) -> None:
        """Prints out a human-readable description of the Airflow integration."""
        print("==================== Airflow Integration  =============================")
        self._metadata.describe()
