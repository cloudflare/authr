<?php declare(strict_types=1);

namespace Cloudflare\Authr;

interface ResourceInterface
{
    /**
     * Get the resource's type slug.
     *
     * @return string
     */
    public function getResourceType(): string;

    /**
     * Get the value for an attribute of the resource.
     *
     * @param string $key
     * @return mixed
     */
    public function getResourceAttribute(string $key);
}
