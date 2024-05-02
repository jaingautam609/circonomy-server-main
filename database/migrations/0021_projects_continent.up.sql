CREATE TYPE continent_values AS ENUM ('asia', 'europe', 'north-america', 'south-america', 'australia', 'africa');

ALTER TABLE IF EXISTS projects ADD COLUMN continent continent_values;
