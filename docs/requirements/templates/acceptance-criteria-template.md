# Acceptance criteria (Gherkin)

Use **one Feature per requirement area** (or per REQ if you prefer strict 1:1). Every **Scenario** MUST declare its parent requirement with the line `Parent requirement: REQ-xxx` immediately under the scenario title (before steps).

Keywords: **Feature**, **Scenario** (or **Scenario Outline** + **Examples**), **Given**, **When**, **Then**, **And**, **But**. Use consistent indentation (two spaces for steps). One blank line between scenarios.

---

## Template

```gherkin
Feature: <Short feature name>

  # Optional: background shared by all scenarios in this feature
  # Background:
  #   Given <precondition>

  Scenario: <Scenario title in plain language>
    Parent requirement: REQ-000
    Given <precondition>
    And <optional additional given>
    When <action>
    Then <expected outcome>
    And <optional additional outcome>

  Scenario Outline: <Parameterized scenario title>
    Parent requirement: REQ-000
    Given <precondition with <placeholder>>
    When <action with <placeholder>>
    Then <outcome with <placeholder>>

    Examples:
      | placeholder | expected |
      | value1      | result1  |
      | value2      | result2  |
```

---

## Rules

- Do not omit `Parent requirement: REQ-xxx` on any scenario.
- Keep scenarios **independent** where possible; repeat Given steps instead of relying on implicit order unless a **Background** is documented for the feature.
- Prefer concrete examples over vague terms (“valid input”, “user sees success” → specify field values and visible text or URLs).
