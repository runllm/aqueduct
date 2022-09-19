import base64

from aqueduct_executor.operators.connectors.data import execute
from aqueduct_executor.operators.connectors.data.spec import parse_spec


def handler(event, context):
    print(event)
    input_spec = event["Spec"]

    spec_json = base64.b64decode(input_spec)
    spec = parse_spec(spec_json)

    execute.run(spec)
