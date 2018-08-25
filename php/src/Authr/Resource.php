<?php declare(strict_types=1);

namespace Cloudflare\Authr;

class Resource implements ResourceInterface
{
    /**
     * The type of the ad-hoc resource
     * 
     * @var string
     */
    private $type;

    /**
     * The attributes of the ad-hoc resource
     * 
     * @var array
     */
    private $attributes = [];

    /**
     * Create a ResourceInterface compatible object that represents an abstract
     * resource.
     *
     * @param string $type
     * @param array $attributes
     * 
     * @return \Cloudflare\Authr\Resource
     * @suppress PhanUnreferencedMethod
     */
    public static function adhoc($type, array $attributes): ResourceInterface
    {
        $rsrc = new static();

        if (!is_string($type) || empty($type)) {
            throw new Exception\InvalidAdHocResourceException('AdHoc resource expects a non-empty string for its type');
        }

        if (!is_array($attributes)) {
            throw new Exception\InvalidAdHocResourceException('AdHoc resource expects an array for its attributes');
        }

        $rsrc->type = $type;
        $rsrc->attributes = $attributes;

        return $rsrc;
    }

    public function getResourceType(): string
    {
        return $this->type;
    }

    public function getResourceAttribute(string $key)
    {
        if (!array_key_exists($key, $this->attributes)) {
            return null;
        }
        $value = $this->attributes[$key];
        if (is_callable($value)) {
            return call_user_func($value);
        }

        return $value;
    }
}
