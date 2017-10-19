<?php

namespace Cloudflare\Authr;

interface ResourceInterface
{
    /**
     * Get the resource's type slug.
     *
     * @return string
     */
    public function getResourceType();

    /**
     * Get the value for an attribute of the resource.
     *
     * @param string $key
     *
     * @return mixed
     */
    public function getResourceAttribute($key);
}
