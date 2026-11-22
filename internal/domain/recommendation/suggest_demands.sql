WITH matches AS (

    -- 1. Busca as demandas e as áreas do professor.
    SELECT
        d.id AS demand_id,
        d.title,
        a.id AS area_id,
        a.name AS area_name,
        pa.type AS area_type,

        -- 2. Calcula a similaridade entre os embeddings.
        GREATEST(0, 1 - (d.embedding <=> a.embedding)) *
        CASE pa.type
            WHEN 'técnica' THEN 0.8
            WHEN 'interesse' THEN 0.5
        END AS score

    FROM demandas d

    JOIN professor_areas pa
        ON pa.professor_id = @professor_id
        AND pa.deleted_at IS NULL

    JOIN areas a
        ON a.id = pa.areas_id

    -- 3. Considera somente demandas disponíveis e com embeddings.
    WHERE
        d.deleted_at IS NULL
        AND d.professor_id IS NULL
        AND d.is_cancelled = false
        AND LOWER(TRIM(d.status)) = 'cadastrada'
        AND d.embedding IS NOT NULL
        AND a.embedding IS NOT NULL
        AND pa.type IN ('técnica', 'interesse')
),

-- 4. Ordena as áreas correspondentes de cada demanda.
ranked AS (
    SELECT
        *,
        ROW_NUMBER() OVER (
            PARTITION BY demand_id
            ORDER BY score DESC, area_id ASC
        ) AS position
    FROM matches
)

-- 5. Retorna somente a melhor correspondência de cada demanda.
SELECT
    demand_id,
    title,
    score,
    area_id,
    area_name,
    area_type
FROM ranked
WHERE position = 1

-- 6. Prioriza as demandas com maior compatibilidade.
ORDER BY score DESC, demand_id ASC
LIMIT @limit
