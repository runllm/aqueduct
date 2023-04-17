import json
import uuid
from abc import ABC
from typing import Any

from pydantic import BaseModel

from aqueduct.constants.enums import ServiceType


class IntegrationInfo(BaseModel):
    id: uuid.UUID
    name: str
    service: ServiceType
    createdAt: int
    validated: bool

    def describe(self) -> None:
        """Prints out a human-readable description of the integration."""
        description_map = {
            "Id": str(self.id),
            "Name": self.name,
            "Service": self.service,
            "CreatedAt": self.createdAt,
            "Validated": self.validated,
        }
        print(json.dumps(description_map, sort_keys=False, indent=4))

    def is_relational(self) -> bool:
        """Returns whether the integration connects to a relational data store."""
        return self.service in [
            ServiceType.POSTGRES,
            ServiceType.SNOWFLAKE,
            ServiceType.MYSQL,
            ServiceType.REDSHIFT,
            ServiceType.MARIADB,
            ServiceType.SQLSERVER,
            ServiceType.BIGQUERY,
            ServiceType.AQUEDUCTDEMO,
            ServiceType.SQLITE,
            ServiceType.ATHENA,
        ]


class Integration(ABC):
    """
    Base class for the various integrations Aqueduct interacts with.
    """

    _metadata: IntegrationInfo

    def id(self) -> uuid.UUID:
        return self._metadata.id

    def name(self) -> str:
        return self._metadata.name

    def type(self) -> ServiceType:
        return self._metadata.service

    def __hash__(self) -> int:
        """An integration is uniquely identified by its name.
        Ref: https://docs.python.org/3.5/reference/datamodel.html#object.__hash__
        """
        return hash(self._metadata.name)

    def __eq__(self, other: Any) -> bool:
        """The string and Integration object representation are equivalent allowing
        the user to access a dictionary keyed by the Integration object with the
        integration name as a string and vice versa
        """
        if type(other) == type(self) and "name" in other._metadata.__dict__:
            return bool(self._metadata.name == other._metadata.name)
        elif type(other) == str:
            return bool(self._metadata.name == other)
        return False
