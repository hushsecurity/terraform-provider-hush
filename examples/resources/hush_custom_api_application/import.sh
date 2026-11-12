# Import by the custom app's name, which also reads its tools. The secret is not
# returned by the API, so the next apply sends it again.
terraform import hush_custom_api_application.billing billing
