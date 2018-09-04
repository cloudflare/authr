import AuthrError from './authrError';
import { isPlainObject, $authr } from './util';

enum Mode {
    BLACKLIST = 0,
    WHITELIST = 1,
    WILDCARD = 2
};

const NOT = '$not';

interface ISlugSetInternal {
    mode: Mode,
    items: string[]
}

interface IBlacklistSpec {
    [NOT]: any
}

function isBlacklistSpec(v: any): v is IBlacklistSpec {
    if (isPlainObject(v)) {
        return v.hasOwnProperty(NOT);
    }
    return false;
}

export default class SlugSet {

    private [$authr]: ISlugSetInternal;

    constructor(spec: any) {
        this[$authr] = {
            mode: Mode.WHITELIST,
            items: []
        };
        if (spec === '*') {
            this[$authr].mode = Mode.WILDCARD;
        } else {
            if (isBlacklistSpec(spec)) {
                this[$authr].mode = Mode.BLACKLIST;
                spec = spec[NOT];
            }
            if (typeof spec === 'string') {
                spec = [spec];
            }
            if (!Array.isArray(spec)) {
                throw new AuthrError('SlugSet constructor expects a string, array or object for argument 1');
            }
            this[$authr].items = spec;
        }
    }

    contains(needle: string): boolean {
        if (this[$authr].mode === Mode.WILDCARD) {
            return true;
        }
        const doesContain = this[$authr].items.includes(needle);
        if (this[$authr].mode === Mode.BLACKLIST) {
            return !doesContain;
        }

        return doesContain;
    }

    toJSON() {
        if (this[$authr].mode === Mode.WILDCARD) {
            return '*';
        }
        let set: any = this[$authr].items;
        if (set.length === 1) {
            [set] = set;
        }
        if (this[$authr].mode === Mode.BLACKLIST) {
            set = { [NOT]: set };
        }
        return set;
    }
}
