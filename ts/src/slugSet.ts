import AuthrError from "./authrError";
import { $authr, isPlainObject } from "./util";

enum Mode {
  BLOCKLIST = 0,
  ALLOWLIST = 1,
  WILDCARD = 2,
}

const NOT = "$not";

interface ISlugSetInternal {
  mode: Mode;
  items: string[];
}

interface IBlocklistSpec {
  [NOT]: any;
}

function isBlocklistSpec(v: any): v is IBlocklistSpec {
  if (isPlainObject(v)) {
    return v.hasOwnProperty(NOT);
  }
  return false;
}

export default class SlugSet {
  private [$authr]: ISlugSetInternal;

  constructor(spec: any) {
    this[$authr] = {
      mode: Mode.ALLOWLIST,
      items: [],
    };
    if (spec === "*") {
      this[$authr].mode = Mode.WILDCARD;
    } else {
      if (isBlocklistSpec(spec)) {
        this[$authr].mode = Mode.BLOCKLIST;
        spec = spec[NOT];
      }
      if (typeof spec === "string") {
        spec = [spec];
      }
      if (!Array.isArray(spec)) {
        throw new AuthrError(
          "SlugSet constructor expects a string, array or object for argument 1"
        );
      }
      this[$authr].items = spec;
    }
  }

  contains(needle: string): boolean {
    if (this[$authr].mode === Mode.WILDCARD) {
      return true;
    }
    const doesContain = this[$authr].items.includes(needle);
    if (this[$authr].mode === Mode.BLOCKLIST) {
      return !doesContain;
    }

    return doesContain;
  }

  toJSON() {
    if (this[$authr].mode === Mode.WILDCARD) {
      return "*";
    }
    let set: any = this[$authr].items;
    if (set.length === 1) {
      [set] = set;
    }
    if (this[$authr].mode === Mode.BLOCKLIST) {
      set = { [NOT]: set };
    }
    return set;
  }
}
