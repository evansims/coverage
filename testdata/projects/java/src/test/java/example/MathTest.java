package example;

import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

class MathTest {
    @Test
    void testAdd() {
        assertEquals(5, Math.add(2, 3));
    }

    @Test
    void testIsPositive() {
        assertTrue(Math.isPositive(1));
        assertFalse(Math.isPositive(-1));
    }
}
