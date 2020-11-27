# authr

[![GO Build Status](https://github.com/cloudflare/authr/workflows/Golang%20Tests/badge.svg)](https://github.com/cloudflare/authr/actions?query=workflow%3A%22Golang+Tests%22)
[![JS Build Status](https://github.com/cloudflare/authr/workflows/JavaScript%20Tests/badge.svg)](https://github.com/cloudflare/authr/actions?query=workflow%3A%22JavaScript+Tests%22)
[![PHP Build Status](https://github.com/cloudflare/authr/workflows/PHP%20Tests/badge.svg)](https://github.com/cloudflare/authr/actions?query=workflow%3A%22PHP+Tests%22)

a flexible, expressive, language-agnostic access-control framework.

## how it works

_authr_ is an access-control framework. describing it as a "framework" is intentional because out of the box it is not going to automatically start securing your application. it is _extremely_ agnostic about quite a few things. it represents building blocks that can be orchestrated and put together in order to underpin an access-control system. by being so fundamental, it can fit almost any need when it comes to controlling access to specific resources in a particular system.

### vocabulary

the framework itself has similar vocabulary to an [ABAC](https://en.wikipedia.org/wiki/Attribute-based_access_control) access-control system. the key terms are explained below.

#### subject

a _subject_ in this framework represents an entity that is capable of performing actions; an _actor_ if you will. in most cases this will represent a "user" or an "admin".

#### resource

a _resource_ represents an entity which can be acted upon. in a blogging application this might be a "post" or a "comment". those are things which can be acted upon by subjects wanting to "edit" them or "delete" them. it _is_ worth noting that subjects can also be resources — a "user" is something that can act and be acted upon.

a _resource_ has **attributes** which can be analyzed by authr. for example, a `post` might have an attribute `id` which is `333`. or, a user might have an attribute `email` which would be `person@awesome.blog`.

#### action

an _action_ is a simple, terse description of what action is being attempted. if say a "user" was attempting to fix a typo in their "post", the _action_ might just be `edit`.

#### rule

a rule is a statement that composes conditions on resource and actions and specifies whether to allow or deny the attempt if the rule is matched. so, for example if you wanted to "allow" a subject to edit a private post, the JSON representation of the rule might look like this:

```json
{
  "access": "allow",
  "where": {
    "action": "edit",
    "rsrc_type": "post",
    "rsrc_match": [["@type", "=", "private"]]
  }
}
```

notice the lack of anything that specifies conditions on _who_ is actually performing the action. this is important; more on that in a second.

### agnosticism through interfaces

across implementations, _authr_ requires that objects implement certain functionality so that its engine can properly analyze resources against a list of rules that _belong_ to a subject.

once the essential objects in an application have implemented these interfaces, the essential question can finally be asked: **can this subject perform this action on this resource?**

```php
<?php

use Cloudflare\Authr;

class UserController extends Controller
{
    /** @var \Cloudflare\AuthrInterface */
    private $authr;
    ...

    public function update(Request $req, Response $res, array $args)
    {
        // get the subject
        $subject = $req->getActor();

        // get the resource
        $resource = $this->getUser($args['id']);

        // check permissions!
        if (!$this->authr->can($subject, 'update', $resource)) {
            throw new HTTPException\Forbidden('Permission denied!');
        }

        ...
    }
}
```

### forming the subject

_authr_ is most of the time identifiable as an ABAC framework. it relies on the ability to place certain conditions on the attributes of resources. there is however one _key_ difference: **there is no way to specify conditions on the subject in rule statements.**

instead, the only way to specify that a specific actor is able to perform an action on a resource is to emit a rule from the returned list of rules that will match the action and allow it to happen. therefore, **a subject is only ever known as a list of rules.**

```go
type Subject interface {
    GetRules() ([]*Rule, error)
}
```

and instead of the rules being statically defined somewhere and needing to make the framework worry about where to retrieve the rules from, **rules belong to subjects** and are only ever retrieved from the subject.

when permissions are checked, the framework will simply call a method available via an interface on the subject to retrieve a list of rules for that specific subject. then, it will iterate through that list until it matches a rule and return a boolean based on whether the rule wanted to allow or deny.

#### why disallow inspection of attributes on the actor?

by reducing actors to just a list of rules, it condenses all of the logic about what a subject is capable of to a single area and keeps it from being scattered all over an application's codebase.

also, in traditional RBAC access-control systems, the notion of checking if a particular actor is in a certain "role" or checking the actors ID to determine access is incredibly brittle and "ages" a codebase.

by having a single component which is responsible for answering the question of access-control, combined with being forced to clearly express what an actor can do with the authr rules, it leads to an incredible separation of concerns and a much more sustainable codebase.

even if authr is not the access-control you choose, there is a distinct advantage to organizing access-control in your services this way, and authr makes sure that things stay that way.

### expressing permissions across service boundaries

because the basic unit of permission in authr is a rule defined in JSON, it is possible to let other services do the access-control checks for their own purposes.

an example of this internally at Cloudflare is in a administrative service. by having this permissions defined in JSON, we can simply transfer all the rules down to the front-end (in JavaScript) and allow the front-end to hide/show certain functionality _based_ on the permission of whoever is logged in.

when you can have the front-end and the back-end of a service seamlessly agreeing with each other on access-control by only updating a single rule, once, it can lead to much easier maintainability.

## todo

- [ ] create integration tests that ensure implementations agree with each other
- [ ] finish go implementation
- [ ] add examples of full apps using authr for access-contro
- [ ] add documentation about the rules format
