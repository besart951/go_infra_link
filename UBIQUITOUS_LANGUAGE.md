# Ubiquitous Language

This glossary is derived from the backend domain model, project handlers, and the alarm model concept. It uses backend-facing names as canonical terms and calls out terms that are currently overloaded.

## Project lifecycle

| Term | Definition | Aliases to avoid |
| --- | --- | --- |
| **Project** | A scoped work context with a name, status, phase, creator, members, and assigned facility objects. | Job, order, site |
| **Project Status** | The lifecycle state of a **Project**: planned, ongoing, or completed. | State, progress |
| **Phase** | A named planning or execution stage that a **Project** belongs to. | Milestone, step |
| **Project Creator** | The **User** that created the **Project**. | Owner, author |
| **Project Member** | A **User** invited to a **Project**. | Collaborator, project user |
| **Project Assignment** | A project-specific association to a facility object such as a **Control Cabinet**, **SPS Controller**, or **Field Device**. | Link, membership, relation |
| **Project Facility Copy** | A copied facility object created for a **Project** from an existing facility object. | Clone, duplicate, template instance |

## Facility topology

| Term | Definition | Aliases to avoid |
| --- | --- | --- |
| **Building** | A physical building identified by an IWS code and building group. | Site, property |
| **Control Cabinet** | A cabinet within a **Building** that contains one or more **SPS Controllers**. | Cabinet, panel |
| **SPS Controller** | A programmable controller installed in a **Control Cabinet** with device identity and network fields. | Controller, PLC, device |
| **SPS Controller System Type** | The pairing of one **SPS Controller** with one **System Type**, optionally numbered and documented. | Controller type, system type link |
| **System Type** | A numbered range that classifies technical systems, such as ventilation or room automation. | Type, category |
| **System Part** | A functional part of a system, identified by short name and full name. | Part, component |
| **Apparat** | A catalogued equipment/function label that can be valid for one or more **System Parts**. | Equipment, apparatus, device type |
| **Field Device** | A projectable facility instance under an **SPS Controller System Type**, classified by **System Part** and **Apparat**. | Device, datapoint carrier |
| **BMK** | The equipment tag used to identify a **Field Device**. | Tag, label, identifier |
| **Specification** | Optional technical product and electrical details for exactly one **Field Device**. | Specs, product data |

## Templates and BACnet

| Term | Definition | Aliases to avoid |
| --- | --- | --- |
| **ObjectData** | A reusable or project-scoped template that groups **BACnet Objects** and eligible **Apparats**. | Object template, object data record |
| **ObjectData Template** | An **ObjectData** record with no **Project**. | Global object data, master data |
| **Project ObjectData** | An **ObjectData** record scoped to exactly one **Project**. | Project template, local object data |
| **BACnet Object** | A BACnet point definition that can belong to **ObjectData** or to a copied **Field Device**. | Point, datapoint, BACnet row |
| **Text Fix** | The stable text identifier of a **BACnet Object**. | Fixed text, label |
| **Text Individual** | Optional custom text that overrides or extends the **Text Fix** for a **BACnet Object**. | Custom text, individual text |
| **Software Type** | The BACnet software object type such as AI, AO, BI, BO, or AV. | Software kind, object type |
| **Software Number** | The number paired with a **Software Type** to identify a software BACnet object. | Object number, software index |
| **Hardware Type** | The physical IO type such as AI, AO, DI, or DO. | IO type, hardware kind |
| **Hardware Quantity** | The count of hardware points represented by a **BACnet Object**. | Quantity, count |
| **Software Reference** | A reference from one **BACnet Object** to another **BACnet Object** used as its software source. | Reference object, linked point |
| **State Text** | A numbered set of up to 16 state labels for BACnet multi-state values. | State labels, text table |
| **Notification Class** | BACnet notification routing and acknowledgement metadata. | Alarm class, NC |
| **GMS Visible** | A flag that marks a **BACnet Object** as visible to the building management system. | Visible, shown in GMS |

## Alarm model

| Term | Definition | Aliases to avoid |
| --- | --- | --- |
| **Alarm Definition** | A named alarm template selected by a **BACnet Object**. | Alarm, alarm config |
| **Alarm Type** | A technical alarm category, such as limit_high_low or io_monitoring, that defines the field schema. | Alarm kind, alarm definition type |
| **Alarm Field** | A global catalog entry for one alarm input, with key, label, data type, and default unit. | Field, alarm input |
| **Alarm Type Field** | The assignment of an **Alarm Field** to an **Alarm Type**, including required, editable, default, validation, and UI grouping rules. | Field rule, alarm field mapping |
| **Alarm Definition Field Override** | A definition-specific override of an **Alarm Type Field**. | Field override, definition override |
| **BACnet Object Alarm Value** | A concrete alarm value stored for one copied **BACnet Object** and one **Alarm Type Field**. | Alarm value, instance value |
| **Unit** | A catalogued measurement unit with code, symbol, and name. | Measurement unit, unit code |
| **Alarm Value Source** | The origin of a **BACnet Object Alarm Value**: default, user, or import. | Source, value origin |

## People and authorization

| Term | Definition | Aliases to avoid |
| --- | --- | --- |
| **User** | An authentication identity that can create projects, belong to teams, and receive roles. | Account, login |
| **Business Details** | Company information attached to exactly one **User**. | Company profile, billing details |
| **Team** | A named group of **Users**. | Group, organization |
| **Team Member** | A **User** in a **Team** with a team-scoped role. | Membership, teammate |
| **Global Role** | A system-wide user role such as superadmin, admin_fzag, planer, or entrepreneur. | User role, account role |
| **Team Member Role** | A team-scoped role: owner, manager, or member. | Team role, member level |
| **Permission** | A named resource-action capability such as project.read or spscontroller.update. | Capability, right |
| **Role Permission** | The assignment of one **Permission** to one **Global Role**. | Permission link, role capability |

## Relationships

- A **Project** belongs to exactly one **Phase** and has exactly one **Project Creator**.
- A **Project** can have many **Project Members**.
- A **Project Assignment** references exactly one **Project** and exactly one assigned facility object.
- A **Building** has many **Control Cabinets**.
- A **Control Cabinet** has many **SPS Controllers**.
- An **SPS Controller** has many **SPS Controller System Types**.
- An **SPS Controller System Type** has many **Field Devices**.
- A **Field Device** belongs to exactly one **SPS Controller System Type**, one **System Part**, and one **Apparat**.
- A **Field Device** has zero or one **Specification**.
- An **ObjectData** groups many **BACnet Objects** and can be valid for many **Apparats**.
- An **ObjectData Template** can be copied into **Field Devices** by cloning its **BACnet Objects**.
- A **BACnet Object** can reference one **State Text**, one **Notification Class**, and one **Alarm Definition**.
- An **Alarm Definition** belongs to zero or one **Alarm Type** in the current code, but the alarm concept intends this to be required.
- An **Alarm Type** has many **Alarm Type Fields**.
- A **BACnet Object Alarm Value** belongs to exactly one copied **BACnet Object** and exactly one **Alarm Type Field**.
- A **Team Member** belongs to exactly one **Team** and exactly one **User**.
- A **Role Permission** belongs to exactly one **Global Role** and one **Permission**.

## Example dialogue

> **Dev:** "When a **Project** needs a ventilation controller, should I attach the existing **SPS Controller** or copy it?"

> **Domain expert:** "Use a **Project Assignment** when the project references an existing object; use a **Project Facility Copy** when the project needs its own editable **SPS Controller**."

> **Dev:** "For **Field Devices**, do I start from **ObjectData**?"

> **Domain expert:** "Yes. Pick an **ObjectData Template**, create the **Field Device**, and clone the template's **BACnet Objects** onto that field device."

> **Dev:** "If a copied **BACnet Object** needs alarm input, do I store fields directly on the object?"

> **Domain expert:** "No. Select an **Alarm Definition**, read its **Alarm Type Fields**, and store concrete **BACnet Object Alarm Values** per copied object."

> **Dev:** "So **Alarm Type** is the schema, **Alarm Definition** is the named template, and **BACnet Object Alarm Value** is the instance data?"

> **Domain expert:** "Exactly. Keep those three separate."

## Flagged ambiguities

- "link" is used in code for project-to-facility associations, but it is too generic for domain conversation; use **Project Assignment** for the business concept and reserve "link" for repository/table implementation details.
- "device" can mean **SPS Controller**, **Field Device**, or a BACnet object-like point; use the full canonical term unless the surrounding aggregate is explicit.
- "type" can mean **System Type**, **Software Type**, **Hardware Type**, **Alarm Type**, or **Global Role**; use qualified names in APIs, DTOs, and discussions.
- "ObjectData" sounds like raw data, but the backend uses it as a reusable **ObjectData Template** or a **Project ObjectData** record; avoid calling it just "data".
- "BACnet Object" can be either template-level or copied onto a **Field Device**; call out whether it is a template BACnet object or a field-device BACnet object when persistence behavior matters.
- "Alarm" is overloaded across **Alarm Definition**, **Alarm Type**, **Alarm Field**, and **BACnet Object Alarm Value**; do not use bare "alarm" in backend names where a specific model exists.
- `BacnetObject.AlarmTypeID` and `BacnetObject.AlarmDefinitionID` both appear in request mapping, while the alarm concept says the BACnet object should select an **Alarm Definition** and derive the **Alarm Type** through it; prefer **Alarm Definition** as the selectable relationship.
- `AlarmDefinition.AlarmTypeID` is nullable in the current model but required in the alarm concept; treat nullable definitions as legacy or incomplete until the migration is finalized.
- "role" can mean **Global Role** or **Team Member Role**; use the qualified term in service boundaries and UI copy.
- "account" should not be used for **User**, **Team**, or **Business Details** because those represent different concepts.
