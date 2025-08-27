<?php
$result = $input |> strtoupper(_) |> trim(_);
$value = $data |> array_filter(_) |> array_values(_);