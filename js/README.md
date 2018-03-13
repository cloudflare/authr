### authr
an incredibly granular and expressive permissions framework.

#### getting started
to get started, install the package!
```
npm install --save @cloudflare/authr
```

in a language that has stronger typing, you might use an interface to ensure your objects have certain functionality. but since javascript doesn't have such a thing, we can use [symbols](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Symbol).

symbols have a lot of uses, but for authr's purposes, it guarantees that a certain object has explicitly implemented certain functionality specifically for authr; kind of like an interface! these symbols are made available for use by the package as exported constants.

*note: a **lot** of important details about the philosophy and inner-workings of this library are glossed over. if you have not read the main README yet, it is highly recommended that you do so: [cloudflare/authr/README.md](https://github.com/cloudflare/authr/blob/master/README.md)*

##### GET_RULES
this constant MUST be the key in an object that is assigned to a callable function. when the function is called, it must return an array of `Rule`s. like this:

```js
import { GET_RULES, Rule } from '@cloudflare/authr';

var user = {
    [GET_RULES]: () => ([
        Rule.allow({
            rsrc_type: 'zone',
            rsrc_match: [['@id', '=', '123']],
            action: ['delete', 'pause']
        }),
        Rule.deny({
            rsrc_type: 'private_key',
            rsrc_match: [],
            action: '*'
        })
    ])
};
```

once you have this, you have just setup a subject in the authr framework! subjects are entities in a system that are capable of altering and affecting resources.

##### GET_RESOURCE_TYPE
this constant MUST be the key in an object that is assigned to a callable function that returns a string that identifies the resource type. like so:

```js
import { GET_RESOURCE_TYPE } from '@cloudflare/authr';

var resource = {
    [GET_RESOURCE_TYPE]: () => 'zone'
};
```

pretty simple function, right? resources in the permission framework are entites that are capable of being acted upon, like a zone, or even a user.

##### GET_RESOURCE_ATTRIBUTE
this constant MUST be the key in an object that is assigned to a callable function. when the function is called, it will be given one string parameter that designates the key of the attribute being looked for. like so:

```js
import {
    GET_RESOURCE_TYPE,
    GET_RESOURCE_ATTRIBUTE
} from '@cloudflare/authr';

var resource = {
    [GET_RESOURCE_TYPE]: () => 'zone',
    [GET_RESOURCE_ATTRIBUTE]: k => {
        var attr = {
            'id': 123,
            'type': 'full'
        };
        return attr[k] || null;
    }
};
```

when using these two constants in an object, you have just setup a resource in the authr framework!

###### but wait! symbols are new as hell and can't be used in old browsers!
mm. true. but there are some decent [shims](https://github.com/medikoo/es6-symbol) that can achieve the same things. not ideal, but, it gets the job done.

#### api
when you have all the objects in place and ready with their constants, you can finally start checking access-control. in this framework, there is only one way to do that: `Authr.can`.

##### Authr.can([subject], [action], [resource])
check if a `subject` is allowed to perform `action` on a particular `resource`.

- `subject` object - the subject attempting the action
- `action` string - the action being attempted
- `resource` object - the resource that will be affected by the change

returns `true` if they are allowed, and `false` if they are not.

###### example
```js
import {
    can
    GET_RULES,
    GET_RESOURCE_TYPE,
    GET_RESOURCE_ATTRIBUTE,
    Rule
} from '@cloudflare/authr';

var admin = {
    [GET_RULES]: () => ([
        Rule.allow({
            rsrc_type: 'zone',
            rsrc_match: [['@status', '=', 'V']],
            action: '*'
        })
    ])
};

var user = {
    [GET_RULES]: () => ([
        Rule.allow({
            rsrc_type: 'zone',
            rsrc_match: [['@id', '=', '123']],
            action: ['init', 'delete']
        })
    ])
};

var zones = {
    '123': {
        [GET_RESOURCE_TYPE]: () => 'zone',
        [GET_RESOURCE_ATTRIBUTE]: k => {
            var attr = {
                'id': 123,
                'status': 'V'
            };
            return attr[k] || null;
        }
    },
    '567': {
        [GET_RESOURCE_TYPE]: () => 'zone',
        [GET_RESOURCE_ATTRIBUTE]: k => {
            var attr = {
                'id': 567,
                'status': 'V'
            };
            return attr[k] || null;
        }
    }
};

console.log(can(admin, 'delete', zones['123'])); // => true
console.log(can(user, 'delete', zones['567'])); // => false
console.log(can(admin, 'do_admin_things', zones['123'])); // true

```
