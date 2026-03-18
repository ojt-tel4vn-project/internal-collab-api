-- ============================================================================
-- MIGRATION: Update Attendance Data Structure
-- Date: 2025-03-18
-- Description: Update attendance_data JSONB to store detailed check-in/out info
-- ============================================================================

-- Previous format: {"1": "present", "2": "absent", ...}
-- New format: {"1": {"status": "present", "check_in_time": "08:19:35", "check_out_time": "19:26:14", "work_hours": 11.0}, ...}

-- ============================================================================
-- 1. Update trigger function to calculate totals from new structure
-- ============================================================================

CREATE OR REPLACE FUNCTION calculate_attendance_totals()
RETURNS TRIGGER AS $$
DECLARE
    day_key TEXT;
    day_detail JSONB;
    status_value TEXT;
    present_count INT := 0;
    absent_count INT := 0;
    late_count INT := 0;
BEGIN
    -- Loop through JSONB object
    FOR day_key IN SELECT jsonb_object_keys(NEW.attendance_data)
    LOOP
        day_detail := NEW.attendance_data->day_key;
        
        -- Check if it's the new format (object) or old format (string)
        IF jsonb_typeof(day_detail) = 'object' THEN
            -- New format: extract status from object
            status_value := day_detail->>'status';
        ELSE
            -- Old format: direct string value
            status_value := day_detail#>>'{}';
        END IF;
        
        -- Count by status
        CASE status_value
            WHEN 'present' THEN present_count := present_count + 1;
            WHEN 'absent' THEN absent_count := absent_count + 1;
            WHEN 'late' THEN late_count := late_count + 1;
        END CASE;
    END LOOP;
    
    NEW.total_days_present := present_count;
    NEW.total_days_absent := absent_count;
    NEW.total_days_late := late_count;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Recreate trigger
DROP TRIGGER IF EXISTS calculate_attendance_totals_trigger ON attendances;
CREATE TRIGGER calculate_attendance_totals_trigger
    BEFORE INSERT OR UPDATE OF attendance_data ON attendances
    FOR EACH ROW EXECUTE FUNCTION calculate_attendance_totals();

-- ============================================================================
-- 2. Optional: Migrate existing data from old format to new format
-- ============================================================================

-- Uncomment below if you want to migrate existing records
/*
DO $$
DECLARE
    rec RECORD;
    day_key TEXT;
    old_status TEXT;
    new_data JSONB := '{}';
BEGIN
    FOR rec IN SELECT id, attendance_data FROM attendances
    LOOP
        new_data := '{}';
        
        -- Loop through each day in the old format
        FOR day_key IN SELECT jsonb_object_keys(rec.attendance_data)
        LOOP
            old_status := rec.attendance_data->>day_key;
            
            -- Check if already in new format
            IF jsonb_typeof(rec.attendance_data->day_key) = 'object' THEN
                new_data := new_data || jsonb_build_object(day_key, rec.attendance_data->day_key);
            ELSE
                -- Convert old format to new format
                new_data := new_data || jsonb_build_object(
                    day_key,
                    jsonb_build_object(
                        'status', old_status,
                        'check_in_time', '',
                        'check_out_time', '',
                        'work_hours', 0
                    )
                );
            END IF;
        END LOOP;
        
        -- Update the record
        UPDATE attendances SET attendance_data = new_data WHERE id = rec.id;
    END LOOP;
END $$;
*/

-- ============================================================================
-- 3. Add comments for documentation
-- ============================================================================

COMMENT ON COLUMN attendances.attendance_data IS 
'JSONB storing daily attendance details. Format: {"day": {"status": "present|absent|late|leave", "check_in_time": "HH:MM:SS", "check_out_time": "HH:MM:SS", "work_hours": 0.0}}';

-- ============================================================================
-- 4. Verification queries
-- ============================================================================

-- Check sample data structure
-- SELECT id, employee_id, month, year, attendance_data FROM attendances LIMIT 5;

-- Check totals calculation
-- SELECT id, total_days_present, total_days_absent, total_days_late FROM attendances LIMIT 5;
