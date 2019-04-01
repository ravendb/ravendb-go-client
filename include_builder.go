package ravendb

/*
type IQueryIncludeBuilder  interface IIncludeBuilder<IQueryIncludeBuilder> {
IQueryIncludeBuilder includeCounter(String path, String name);

IQueryIncludeBuilder includeCounters(String path, String[] names);

IQueryIncludeBuilder includeAllCounters(String path);

TSelf includeCounter(String name);

TSelf includeCounters(String[] names);

TSelf includeAllCounters();

TSelf includeDocuments(String path);
}
*/

type TupleBoolStringSet struct {
	b   bool
	set []string // set of strings
}

type IncludeBuilder struct {
	nextParameterId               int // TODO: default is 1
	_conventions                  *DocumentConventions
	documentsToInclude            map[string]struct{} // set of string
	alias                         string
	countersToIncludeBySourcePath map[string]*TupleBoolStringSet
}

/*
public class IncludeBuilder implements IQueryIncludeBuilder {

    public Set<String> getCountersToInclude() {
        if (countersToIncludeBySourcePath == null) {
            return null;
        }

        Tuple<Boolean, Set<String>> value = countersToIncludeBySourcePath.get("");

        return value != null ? value.second : new HashSet<>();
    }

    public boolean isAllCounters() {
        if (countersToIncludeBySourcePath == null) {
            return false;
        }

        Tuple<Boolean, Set<String>> value = countersToIncludeBySourcePath.get("");
        return value != null ? value.first : false;
    }

    public IncludeBuilder(DocumentConventions conventions) {
        _conventions = conventions;
    }

    @Override
    public IQueryIncludeBuilder includeCounter(String path, String name) {
        _includeCounterWithAlias(path, name);
        return this;
    }

    @Override
    public IQueryIncludeBuilder includeCounters(String path, String[] names) {
        _includeCounterWithAlias(path, names);
        return this;
    }

    @Override
    public IQueryIncludeBuilder includeAllCounters(String path) {
        _includeAllCountersWithAlias(path);
        return this;
    }

    @Override
    public IQueryIncludeBuilder includeCounter(String name) {
        _includeCounter("", name);
        return this;
    }

    @Override
    public IQueryIncludeBuilder includeCounters(String[] names) {
        _includeCounters("", names);
        return this;
    }

    @Override
    public IQueryIncludeBuilder includeAllCounters() {
        _includeAllCounters("");
        return this;
    }

    @Override
    public IQueryIncludeBuilder includeDocuments(String path) {
        _includeDocuments(path);
        return this;
    }

    private void _includeCounterWithAlias(String path, String name) {
        _withAlias();
        _includeCounter(path, name);
    }

    private void _includeCounterWithAlias(String path, String[] names) {
        _withAlias();
        _includeCounters(path, names);
    }

    private void _includeDocuments(String path) {
        if (documentsToInclude == null) {
            documentsToInclude = new HashSet<>();
        }

        documentsToInclude.add(path);
    }

    private void _includeCounter(String path, String name) {
        if (StringUtils.isEmpty(name)) {
            throw new IllegalArgumentException("Name cannot be empty");
        }

        assertNotAllAndAddNewEntryIfNeeded(path);

        countersToIncludeBySourcePath.get(path).second.add(name);
    }

    private void _includeCounters(String path, String[] names) {
        if (names == null) {
            throw new IllegalArgumentException("Names cannot be null");
        }

        assertNotAllAndAddNewEntryIfNeeded(path);

        for (String name : names) {
            if (StringUtils.isWhitespace(name)) {
                throw new IllegalArgumentException("Counters(String[] names): 'names' should not contain null or whitespace elements");
            }

            countersToIncludeBySourcePath.get(path).second.add(name);
        }
    }

    private void _includeAllCountersWithAlias(String path) {
        _withAlias();
        _includeAllCounters(path);
    }

    private void _includeAllCounters(String sourcePath) {
        if (countersToIncludeBySourcePath == null) {
            countersToIncludeBySourcePath = new TreeMap<>(String::compareToIgnoreCase);
        }

        Tuple<Boolean, Set<String>> val = countersToIncludeBySourcePath.get(sourcePath);

        if (val != null && val.second != null) {
            throw new IllegalStateException("You cannot use allCounters() after using counter(String name) or counters(String[] names)");
        }

        countersToIncludeBySourcePath.put(sourcePath, Tuple.create(true, null));
    }

    private void assertNotAllAndAddNewEntryIfNeeded(String path) {
        if (countersToIncludeBySourcePath != null) {
            Tuple<Boolean, Set<String>> val = countersToIncludeBySourcePath.get(path);
            if (val != null && val.first) {
                throw new IllegalStateException("You cannot use counter(name) after using allCounters()");
            }
        }

        if (countersToIncludeBySourcePath == null) {
            countersToIncludeBySourcePath = new TreeMap<>(String::compareToIgnoreCase);
        }

        if (!countersToIncludeBySourcePath.containsKey(path)) {
            countersToIncludeBySourcePath.put(path, Tuple.create(false, new TreeSet<>(String::compareToIgnoreCase)));
        }
    }

    private void _withAlias() {
        if (alias == null) {
            alias = "a_" + (nextParameterId++);
        }
    }
}
*/
