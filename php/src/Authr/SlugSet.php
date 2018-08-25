<?php declare(strict_types=1);

namespace Cloudflare\Authr;

use JsonSerializable;

final class SlugSet implements JsonSerializable
{
    const MODE_BLACKLIST = 0;
    const MODE_WHITELIST = 1;
    const MODE_WILDCARD = 2;

    const NOT = '$not';

    private $mode = self::MODE_WHITELIST;

    /** @var string[] */
    private $items = [];

    /**
     * @param string|array $spec
     * @throws \Cloudflare\Authr\Exception\InvalidSlugSetException
     */
    public function __construct($spec)
    {
        if ($spec === '*') {
            $this->mode = static::MODE_WILDCARD;
        } else {
            if (is_array($spec) && key($spec) === static::NOT) {
                $this->mode = static::MODE_BLACKLIST;
                $spec = $spec[static::NOT];
            }
            if (is_string($spec)) {
                $spec = [$spec];
            }
            if (!is_array($spec)) {
                throw new Exception\InvalidSlugSetException('SlugSet constructor expects a string or an array for argument 1');
            }
            $this->items = $spec;
        }
    }

    /**
     * @param string $needle
     * @return bool
     */
    public function contains(string $needle): bool
    {
        if ($this->mode === static::MODE_WILDCARD) {
            return true;
        }
        $doesContain = in_array($needle, $this->items, true);
        if ($this->mode === static::MODE_BLACKLIST) {
            return !$doesContain;
        }

        return $doesContain;
    }

    /**
     * @return mixed
     */
    public function jsonSerialize()
    {
        if ($this->mode === static::MODE_WILDCARD) {
            return '*';
        }
        $set = $this->items;
        if (count($set) === 1) {
            $set = $set[0];
        }
        if ($this->mode === static::MODE_BLACKLIST) {
            $set = [static::NOT => $set];
        }

        return $set;
    }
}
