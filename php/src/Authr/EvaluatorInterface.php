<?php declare(strict_types=1);

namespace Cloudflare\Authr;

interface EvaluatorInterface extends \JsonSerializable
{
    /**
     * Evaluate something against a resource
     * 
     * @param \Cloudflare\Authr\ResourceInterface $resource
     * @return bool
     */
    public function evaluate(ResourceInterface $resource): bool;
}
