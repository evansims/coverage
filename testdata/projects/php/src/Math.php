<?php

namespace Example;

class Math
{
    public function add(int $a, int $b): int
    {
        return $a + $b;
    }

    public function isPositive(int $n): bool
    {
        if ($n > 0) {
            return true;
        }
        return false;
    }
}
