# 🧪 Testing Plugin

The **Testing Plugin** provides nodes for testing and validation. By using the `test` command, you can execute defined
nodes in your workflow, handling not only simple success/failure validation but also more complex testing scenarios.
This plugin helps ensure that workflows function as expected during development and allows for early detection and
correction of errors.

## Available Nodes

- **[Test Node](./docs/test_node.md)**: A node that validates the results of workflow execution and determines whether
  the outcome is a success or failure. You can write various validation logic to accurately verify the execution
  results.
- **[Assert Node](./docs/assert_node.md)**: A node that compares expected results with actual execution outcomes to
  verify if the two values match. Typically used with the `Test Node`, it allows for more refined testing by setting
  complex validation conditions.
