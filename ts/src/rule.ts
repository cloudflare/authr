import { $authr, isPlainObject, IJSONSerializable } from './util';
import AuthrError from './authrError';
import SlugSet from './slugSet';
import ConditionSet from './conditionSet';

export enum Access {
    ALLOW = 'allow',
    DENY = 'deny'
}

function coerceToAccess(v: any): Access {
    if (typeof v === 'string') {
        switch (v) {
            case Access.ALLOW:
            case Access.DENY:
                return v;
        }
    }
    throw new AuthrError(`invalid "access" value: "${v}"`);
}

export const RSRC_TYPE = 'rsrc_type';
export const RSRC_MATCH = 'rsrc_match';
export const ACTION = 'action';

interface IRuleInternal {
    access: Access,
    where: {
        [RSRC_TYPE]?: SlugSet,
        [RSRC_MATCH]?: ConditionSet,
        [ACTION]?: SlugSet
    },
    meta: any
}

export default class Rule implements IJSONSerializable {

    private [$authr]: IRuleInternal;

    static allow(spec: any, meta: any = null): Rule {
        return new Rule(Access.ALLOW, spec, meta);
    }

    static deny(spec: any, meta: any = null): Rule {
        return new Rule(Access.DENY, spec, meta);
    }

    static create(spec: any) {
        if (!isPlainObject(spec)) {
            throw new AuthrError('"spec" must be a plain object');
        }
        return new Rule(
            (spec as any).access,
            (spec as any).where,
            (spec as any).$meta
        );
    }

    constructor(access: any, spec: any, meta: any = null) {
        this[$authr] = {
            access: coerceToAccess(access),
            where: {},
            meta: null
        };
        if (typeof spec === 'string' && spec === 'all') {
            spec = {
                [RSRC_TYPE]: '*',
                [RSRC_MATCH]: [],
                [ACTION]: '*'
            };
        }
        if (!isPlainObject(spec)) {
            throw new AuthrError('"spec" must be a string or plain object');
        }
        if (meta) {
            this[$authr].meta = meta;
        }
        for (let seg in spec) {
            let segspec: any = (spec as any)[seg];
            switch (seg) {
                case RSRC_TYPE:
                case ACTION:
                    this[$authr].where[seg] = new SlugSet(segspec);
                    break;
                case RSRC_MATCH:
                    this[$authr].where[RSRC_MATCH] = new ConditionSet(segspec);
                    break;
            }
        }
    }

    access(): Access {
        return this[$authr].access;
    }

    resourceTypes(): SlugSet {
        const rt = this[$authr].where[RSRC_TYPE];
        if (!rt) {
            throw new AuthrError('missing "where.rsrc_type" segment on rule');
        }
        return rt;
    }

    conditions(): ConditionSet {
        const match = this[$authr].where[RSRC_MATCH];
        if (!match) {
            throw new AuthrError('missing "where.rsrc_match" segment on rule');
        }
        return match;
    }

    actions(): SlugSet {
        const act = this[$authr].where[ACTION];
        if (!act) {
            throw new AuthrError('missing "where.action" segment on rule');
        }
        return act;
    }

    toJSON(): any {
        interface IRaw {
            access: string,
            where: {
                [RSRC_TYPE]: any,
                [RSRC_MATCH]: any,
                [ACTION]: any,
            },
            $meta?: any
        }
        const raw: IRaw = {
            access: this[$authr].access,
            where: {
                [RSRC_TYPE]: this.resourceTypes().toJSON(),
                [RSRC_MATCH]: this.conditions().toJSON(),
                [ACTION]: this.actions().toJSON()
            }
        };
        if (this[$authr].meta) {
            raw.$meta = this[$authr].meta;
        }
        return raw;
    }

    toString() {
        return JSON.stringify(this);
    }
}
