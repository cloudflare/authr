<?php declare(strict_types=1);

namespace Cloudflare;

use Cloudflare\Authr\ResourceInterface;
use Cloudflare\Authr\SubjectInterface;

interface AuthrInterface
{
    /**
     * Check if a certain actor is allowed to perform an action against a
     * particular resource.
     *
     * @param \Cloudflare\Authr\SubjectInterface $subject The thing that is performing the action
     * @param string $action
     * @param \Cloudflare\Authr\ResourceInterface $resource
     * @return bool
     */
    public function can(SubjectInterface $subject, string $action, ResourceInterface $resource): bool;

    /**
     * Validate a raw definition of a rule.
     * @param array $definition
     * @return void
     * @throws \Cloudflare\Authr\Exception\ValidationException
     */
    public function validateRule($definition): void;
}
