import re
from datetime import date
from typing import Any, Dict, Optional, Union

from aqueduct.integrations.sql_integration import BUILT_IN_EXPANSIONS, TAG_PATTERN
from aqueduct_executor.operators.connectors.data import common, models
from aqueduct_executor.operators.utils import enums


def replace_today() -> str:
    return "'" + date.today().strftime("%Y-%m-%d") + "'"


class RelationalParams(models.BaseParams):
    # The query cannot be used until `apply_placeholders()` is called on it. This flushes out
    # any user-defined tags like `{{today}}`.
    query_is_usable: Optional[bool] = False

    query: str

    # TODO: Consider not including github as part of relational params when it is JSON marshalled
    github_metadata: Optional[Any]

    def expand_placeholders(
        self,
        parameters: Dict[str, str],
    ) -> None:
        """Expands any tags found in the raw query, eg. {{ today }}.

        Relational queries can be arbitrarily parameterized the same way operators are. The only
        requirement is that these parameters must be defined as strings.

        User-defined parameters are prioritized over built-in ones. Eg. if the user defines a parameter
        named "today" that they set with value "1234", the "{{today}}" will be expanded as "1234", even
        though there already is a built-in expansion.
        """
        orig_query = self.query
        matches = re.findall(TAG_PATTERN, self.query)
        for match in matches:
            tag_name = match.strip(" {}")

            if tag_name in parameters:
                self.query = self.query.replace(match, parameters[tag_name])
            elif tag_name in BUILT_IN_EXPANSIONS:
                expansion_func = BUILT_IN_EXPANSIONS[tag_name]
                self.query = self.query.replace(match, expansion_func())
            else:
                # If there's a tag in the query for which no expansion value is available, we error here.
                raise Exception(
                    "Unable to expand tag `%s` for query `%s`." % (tag_name, orig_query)
                )

        print("Expanded query is `%s`." % self.query)
        self.query_is_usable = True

    def usable(self) -> bool:
        """Denotes whether all placeholders have already been expanded for this query.

        Callers should check that `usable()` -> True before actually executing this query.
        """
        # We cannot return self.query_is_usable directly, since it is an Optional
        # and the method expects a bool to be returned.
        if self.query_is_usable:
            return True
        return False


class S3Params(models.BaseParams):
    filepath: str
    artifact_type: enums.ArtifactType
    format: Optional[common.S3TableFormat]
    merge: Optional[bool]


Params = Union[RelationalParams, S3Params]
