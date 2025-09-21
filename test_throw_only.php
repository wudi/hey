<?php
echo "before throw\n";
throw new Exception("test error");
echo "after throw - should not see this\n";