package ravendb

type MethodsType = string

const (
	MethodsType_CMP_X_CHG = "CmpXChg"
)

type WhereToken struct {
	*QueryToken
}

func NewWhereToken() *WhereToken {
	return &WhereToken{
		QueryToken: NewQueryToken(),
	}
}

type WhereMethodCall struct {
	methodType MethodsType
	parameters []string
	property   string
}

/*
public class WhereToken extends QueryToken {

    public static class WhereOptions {
        private SearchOperator searchOperator;
        private String fromParameterName;
        private String toParameterName;
        private Double boost;
        private Double fuzzy;
        private Integer proximity;
        private boolean exact;
        private WhereMethodCall method;
        private ShapeToken whereShape;
        private double distanceErrorPct;

        public static WhereOptions defaultOptions() {
            return new WhereOptions();
        }

        private WhereOptions() {
        }

        public WhereOptions(boolean exact) {
            this.exact = exact;
        }

        public WhereOptions(boolean exact, String from, String to) {
            this.exact = exact;
            this.fromParameterName = from;
            this.toParameterName = to;
        }

        public WhereOptions(SearchOperator search) {
            this.searchOperator = search;
        }

        public WhereOptions(ShapeToken shape, double distance) {
            whereShape = shape;
            distanceErrorPct = distance;
        }

        public WhereOptions(MethodsType methodType, String[] parameters, String property) {
            this(methodType, parameters, property, false);
        }

        public WhereOptions(MethodsType methodType, String[] parameters, String property, boolean exact) {
            method = new WhereMethodCall();
            method.methodType = methodType;
            method.parameters = parameters;
            method.property = property;

            this.exact = exact;
        }

        public SearchOperator getSearchOperator() {
            return searchOperator;
        }

        public void setSearchOperator(SearchOperator searchOperator) {
            this.searchOperator = searchOperator;
        }

        public String getFromParameterName() {
            return fromParameterName;
        }

        public void setFromParameterName(String fromParameterName) {
            this.fromParameterName = fromParameterName;
        }

        public String getToParameterName() {
            return toParameterName;
        }

        public void setToParameterName(String toParameterName) {
            this.toParameterName = toParameterName;
        }

        public Double getBoost() {
            return boost;
        }

        public void setBoost(Double boost) {
            this.boost = boost;
        }

        public Double getFuzzy() {
            return fuzzy;
        }

        public void setFuzzy(Double fuzzy) {
            this.fuzzy = fuzzy;
        }

        public Integer getProximity() {
            return proximity;
        }

        public void setProximity(Integer proximity) {
            this.proximity = proximity;
        }

        public boolean isExact() {
            return exact;
        }

        public void setExact(boolean exact) {
            this.exact = exact;
        }

        public WhereMethodCall getMethod() {
            return method;
        }

        public void setMethod(WhereMethodCall method) {
            this.method = method;
        }

        public ShapeToken getWhereShape() {
            return whereShape;
        }

        public void setWhereShape(ShapeToken whereShape) {
            this.whereShape = whereShape;
        }

        public double getDistanceErrorPct() {
            return distanceErrorPct;
        }

        public void setDistanceErrorPct(double distanceErrorPct) {
            this.distanceErrorPct = distanceErrorPct;
        }

    }

    private String fieldName;
    private WhereOperator whereOperator;
    private String parameterName;
    private WhereOptions options;

    public static WhereToken create(WhereOperator op, String fieldName, String parameterName) {
        return create(op, fieldName, parameterName, null);
    }

    public static WhereToken create(WhereOperator op, String fieldName, String parameterName, WhereOptions options) {
        WhereToken token = new WhereToken();
        token.fieldName = fieldName;
        token.parameterName = parameterName;
        token.whereOperator = op;
        token.options = ObjectUtils.firstNonNull(options, WhereOptions.defaultOptions());
        return token;
    }

    public String getFieldName() {
        return fieldName;
    }

    public void setFieldName(String fieldName) {
        this.fieldName = fieldName;
    }

    public WhereOperator getWhereOperator() {
        return whereOperator;
    }

    public void setWhereOperator(WhereOperator whereOperator) {
        this.whereOperator = whereOperator;
    }

    public String getParameterName() {
        return parameterName;
    }

    public void setParameterName(String parameterName) {
        this.parameterName = parameterName;
    }

    public WhereOptions getOptions() {
        return options;
    }

    public void setOptions(WhereOptions options) {
        this.options = options;
    }

    public void addAlias(String alias) {
        if ("id()".equals(fieldName)) {
            return;
        }
        fieldName = alias + "." + fieldName;
    }

    private boolean writeMethod(StringBuilder writer) {
        if (options.getMethod() != null) {
            switch (options.getMethod().methodType) {
                case CMP_X_CHG:
                    writer.append("cmpxchg(");
                    break;
                default:
                    throw new IllegalArgumentException("Unsupported method: " + options.getMethod().methodType);
            }

            boolean first = true;
            for (String parameter : options.getMethod().parameters) {
                if (!first) {
                    writer.append(",");
                }
                first = false;
                writer.append("$");
                writer.append(parameter);
            }
            writer.append(")");

            if (options.getMethod().property != null) {
                writer.append(".")
                        .append(options.getMethod().property);
            }
            return true;
        }

        return false;
    }

    @Override
    public void writeTo(StringBuilder writer) {
        if (options.boost != null) {
            writer.append("boost(");
        }

        if (options.fuzzy != null) {
            writer.append("fuzzy(");
        }

        if (options.proximity != null) {
            writer.append("proximity(");
        }

        if (options.exact) {
            writer.append("exact(");
        }

        switch (whereOperator) {
            case SEARCH:
                writer.append("search(");
                break;
            case LUCENE:
                writer.append("lucene(");
                break;
            case STARTS_WITH:
                writer.append("startsWith(");
                break;
            case ENDS_WITH:
                writer.append("endsWith(");
                break;
            case EXISTS:
                writer.append("exists(");
                break;
            case SPATIAL_WITHIN:
                writer.append("spatial.within(");
                break;
            case SPATIAL_CONTAINS:
                writer.append("spatial.contains(");
                break;
            case SPATIAL_DISJOINT:
                writer.append("spatial.disjoint(");
                break;
            case SPATIAL_INTERSECTS:
                writer.append("spatial.intersects(");
                break;
            case REGEX:
                writer.append("regex(");
                break;
        }

        writeInnerWhere(writer);

        if (options.exact) {
            writer.append(")");
        }

        if (options.proximity != null) {
            writer
                    .append(", ")
                    .append(options.proximity)
                    .append(")");
        }

        if (options.fuzzy != null) {
            writer
                    .append(", ")
                    .append(options.fuzzy)
                    .append(")");
        }

        if (options.boost != null) {
            writer
                    .append(", ")
                    .append(options.boost)
                    .append(")");
        }
    }

    private void writeInnerWhere(StringBuilder writer) {

        writeField(writer, fieldName);

        switch (whereOperator) {
            case EQUALS:
                writer
                        .append(" = ");
                break;

            case NOT_EQUALS:
                writer
                        .append(" != ");
                break;
            case GREATER_THAN:
                writer
                        .append(" > ");
                break;
            case GREATER_THAN_OR_EQUAL:
                writer
                        .append(" >= ");
                break;
            case LESS_THAN:
                writer
                        .append(" < ");
                break;
            case LESS_THAN_OR_EQUAL:
                writer
                        .append(" <= ");
                break;
            default:
                specialOperator(writer);
                return;
        }

        if (!writeMethod(writer)) {
            writer.append("$").append(parameterName);
        }
    }

    private void specialOperator(StringBuilder writer) {
        switch (whereOperator)
        {
            case IN:
                writer
                        .append(" in ($")
                        .append(parameterName)
                        .append(")");
                break;
            case ALL_IN:
                writer
                        .append(" all in ($")
                        .append(parameterName)
                        .append(")");
                break;
            case BETWEEN:
                writer
                        .append(" between $")
                        .append(options.fromParameterName)
                        .append(" and $")
                        .append(options.toParameterName);
                break;

            case SEARCH:
                writer
                        .append(", $")
                        .append(parameterName);
                if (options.searchOperator == SearchOperator.AND) {
                    writer.append(", and");
                }
                writer.append(")");
                break;
            case LUCENE:
            case STARTS_WITH:
            case ENDS_WITH:
            case REGEX:
                writer
                        .append(", $")
                        .append(parameterName)
                        .append(")");
                break;
            case EXISTS:
                writer
                        .append(")");
                break;
            case SPATIAL_WITHIN:
            case SPATIAL_CONTAINS:
            case SPATIAL_DISJOINT:
            case SPATIAL_INTERSECTS:
                writer
                        .append(", ");
                options.whereShape.writeTo(writer);

                if (Math.abs(options.distanceErrorPct - Constants.Documents.Indexing.Spatial.DEFAULT_DISTANCE_ERROR_PCT) > 1e-40) {
                    writer.append(", ");
                    writer.append(options.distanceErrorPct);
                }
                writer
                        .append(")");
                break;
            default:
                throw new IllegalArgumentException();
        }
    }
}
*/
