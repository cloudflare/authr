<?php declare(strict_types=1);

namespace Cloudflare\Authr;

final class Rule implements \JsonSerializable
{
    /**
     * A key in a rule where its value is either "allow" or "deny." This tells
     * the evaluator what to do when a rule is matched.
     */
    const ACCESS = 'access';

    /**
     * A token for a rule that sets it as an allowing rule. That is, when it
     * is matched, it will permit the action being evaluated.
     */
    const ALLOW = 'allow';

    /**
     * A token for a rule that sets it as a denying rule. When the denying
     * rule is matched, it will reject the action being evaluated.
     */
    const DENY = 'deny';

    /**
     * A key in a rule where its value is a map the defines a set of conditions
     * to match in an authorization check.
     */
    const WHERE = 'where';

    /**
     * The key of the segment of a rule's 'where' section that defines resource
     * types to match or not match.
     */
    const RESOURCE_TYPE = 'rsrc_type';

    /**
     * The key of the segment of a rule's 'where' section that defines a set of
     * conditions to match specific resources.
     */
    const RESOURCE_MATCH = 'rsrc_match';

    /**
     * The key of the segment of a rule's 'where' section that defines actions
     * to match or not match.
     */
    const ACTION = 'action';

    /**
     * A key in a rule where its value can be anything. This is useful for
     * storing information about its origins or for distiguishing it somewhow.
     */
    const META = '$meta';

    /**
     * The access of the rule. Do we allow or deny if the rule is matched?
     * Defaults to being an allowing rule.
     * 
     * @var string
     */
    private $access = self::ALLOW;

    /**
     * The rule's "where" clause.
     *
     * @var array
     */
    private $where = [
        self::RESOURCE_TYPE => null,
        self::RESOURCE_MATCH => null,
        self::ACTION => null,
    ];

    /**
     * The rule's metadata.
     * 
     * @var mixed
     */
    private $meta;

    /**
     * Construct an allowing rule
     * 
     * @param array|string $where
     * @param mixed $meta Set any metadata on the rule
     * @return self
     * @suppress PhanUnreferencedMethod
     */
    public static function allow($where, $meta = null): self
    {
        return new static(static::ALLOW, $where, $meta);
    }

    /**
     * Construct a denying rule
     * 
     * @param array|string $where
     * @param mixed $meta Set any metadata on the rule
     * @return static
     * @suppress PhanUnreferencedMethod
     */
    public static function deny($where, $meta = null): self
    {
        return new static(static::DENY, $where, $meta);
    }

    /**
     * Create a rule from its JSON definition or raw array form
     * 
     * @param array|string $spec
     * @return static
     * @throws \Cloudflare\Authr\Exception\InvalidRuleException If the policy definition is invalid
     * @throws \Cloudflare\Authr\Exception\RuntimeException If JSON decoding failed
     * @suppress PhanUnreferencedMethod
     */
    public static function create($spec): self
    {
        if (is_string($spec)) {
            $spec = json_decode($spec, true);
            $err = json_last_error();
            if ($err !== \JSON_ERROR_NONE) {
                throw new Exception\RuntimeException(sprintf('Failed to decode rule as JSON: %s', json_last_error_msg()), json_last_error());
            }
        }
        if (!is_array($spec)) {
            throw new Exception\InvalidRuleException(sprintf('%s::create expects a string or array for argument 1, got %s', static::class, gettype($spec)));
        }

        $meta = null;
        if (array_key_exists(static::ACCESS, $spec)) {
            $access = $spec[static::ACCESS];
            if ($access !== static::ALLOW && $access !== static::DENY) {
                throw new Exception\InvalidRuleException(
                    sprintf(
                        "Rule constructor expects '%s' or '%s' as the only values assigned to '%s', got '%s'",
                        static::ALLOW,
                        static::DENY,
                        static::ACCESS,
                        strval($access)
                    )
                );
            }
        }
        if (array_key_exists(static::WHERE, $spec)) {
            $where = $spec[static::WHERE];
            // only validate type, __construct will make sure the whole section
            // is valid
            if (!is_array($spec)) {
                throw new Exception\InvalidRuleException(sprintf('%s::create expects a map to be assigned to \'%s\', got a %s', static::class, static::WHERE, gettype($where)));
            }
        }
        if (array_key_exists(static::META, $spec)) {
            $meta = $spec[static::META];
        }

        $missingkeys = [];
        if (!isset($access)) {
            $missingkeys[] = static::ACCESS;
        }
        if (!isset($where)) {
            $missingkeys[] = static::WHERE;
        }
        if (!empty($missingkeys)) {
            throw new Exception\InvalidRuleException(sprintf('Rule definition missing map keys: %s', implode(', ', $missingkeys)));
        }

        return new static($access, $where, $meta);
    }

    /**
     * Construct a new rule. Rule MUST be immutable after construction.
     * 
     * @param string $access
     * @param array|string $where
     * @param mixed $meta
     */
    private function __construct($access, $where, $meta)
    {
        $this->access = $access;
        $this->meta = $meta;
        if ($where === 'all') {
            $where = [
                static::RESOURCE_TYPE => '*',
                static::RESOURCE_MATCH => [],
                static::ACTION => '*'
            ];
        }
        if (!is_array($where)) {
            throw new Exception\InvalidRuleException(sprintf("Rule constructor expectes 'all' or array for argument 2, got %s", gettype($where)));
        }

        foreach ($where as $seg => $segspec) {
            switch ($seg) {
                case static::RESOURCE_TYPE:
                case static::ACTION:
                    $this->where[$seg] = new SlugSet($segspec);
                    break;
                case static::RESOURCE_MATCH:
                    $this->where[$seg] = new ConditionSet($segspec);
                    break;
                default:
                    throw new Exception\InvalidRuleException(sprintf("Rule constructor included an unknown map key in the 'where' section: %s", $seg));
            }
        }
        $nullwhere = [];
        foreach ($this->where as $key => $value) {
            if (is_null($value)) {
                $nullwhere[] = $key;
            }
        }
        if (!empty($nullwhere)) {
            throw new Exception\InvalidRuleException(sprintf("Rule constructor is missing key(s) in the where section: %s", implode(', ', $nullwhere)));
        }
    }

    /**
     * Retrieve the rule's access
     * 
     * @return string
     */
    public function access(): string
    {
        return $this->access;
    }

    /**
     * Retrieve the rule's resource type segment
     * 
     * @return \Cloudflare\Authr\SlugSet
     * @throws \Cloudflare\Authr\Exception\RuntimeException If the segment is undefined
     */
    public function resourceTypes(): SlugSet
    {
        if (is_null($this->where[static::RESOURCE_TYPE])) {
            throw new Exception\RuntimeException('Cannot retrieve undefined resource type segment');
        }

        return $this->where[static::RESOURCE_TYPE];
    }

    /**
     * Retrieve the rule's conditions segment
     * 
     * @return \Cloudflare\Authr\ConditionSet
     * @throws \Cloudflare\Authr\Exception\RuntimeException If the segment is undefined
     */
    public function conditions(): ConditionSet
    {
        if (is_null($this->where[static::RESOURCE_MATCH])) {
            throw new Exception\RuntimeException('Cannot retrieve undefined resource match segment');
        }

        return $this->where[static::RESOURCE_MATCH];
    }

    /**
     * Retrieve the rule's action segment
     * 
     * @return \Cloudflare\Authr\SlugSet
     * @throws \Cloudflare\Authr\Exception\RuntimeException If the segment is undefined
     */
    public function actions(): SlugSet
    {
        if (is_null($this->where[static::ACTION])) {
            throw new Exception\RuntimeException('Cannot retrieve undefined actions segment');
        }

        return $this->where[static::ACTION];
    }

    /**
     * Retrieve the rule's metadata
     * 
     * @return mixed
     */
    public function meta()
    {
        return $this->meta;
    }

    /**
     * Decompose the policy to be encoded into JSON
     * 
     * @return array
     */
    public function jsonSerialize()
    {
        $raw = [
            static::ACCESS => $this->access,
            static::WHERE => [
                static::RESOURCE_TYPE => $this->resourceTypes(),
                static::RESOURCE_MATCH => $this->conditions(),
                static::ACTION => $this->actions(),
            ]
        ];
        if (!is_null($this->meta)) {
            $raw[static::META] = $this->meta;
        }
        return $raw;
    }

    /**
     * Stringify the policy into JSON
     * 
     * @return string
     */
    public function __toString()
    {
        return json_encode($this);
    }
}
